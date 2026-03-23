// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package context

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	pgengine "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/engine"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	groupresolver "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/resolver"
	modelretry "github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy/model_retry"
)

// routeToModelInstance is the thin entry used by the request context filter.
// The actual model-instance picking flow lives in this file and the unhealthy
// fallback special case lives in the sibling model_pick_unhealthy.go.
func routeToModelInstance(ctx context.Context, clientID, modelName string, headers http.Header) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	return pickModelInstance(ctx, clientID, modelName, headers, modelPickDeps{
		routeEngine:   pgengine.GetEngine(),
		healthManager: health.GetManager(),
	})
}

type modelPickAttempt struct {
	clientID         string
	headers          http.Header
	group            *policypb.PolicyGroup
	routingInstances []*policygroup.RoutingModelInstance
	meta             policygroup.RequestMeta
	routeEngine      *pgengine.Engine
	healthManager    *health.Manager
	pickedAt         time.Time
}

type modelPickDeps struct {
	routeEngine   *pgengine.Engine
	healthManager *health.Manager
	pickedAt      time.Time
}

// pickModelInstance picks the final model instance for the current request.
// It first applies the policy-group routing rules. Only when that normal pick
// has no available branch/instance during a retry attempt do we fall back to
// the unhealthy-specific rescue path in the sibling file.
func pickModelInstance(
	ctx context.Context,
	clientID, modelName string,
	headers http.Header,
	deps modelPickDeps,
) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	resolver := groupresolver.NewResolver()
	group, err := resolver.Resolve(ctx, clientID, modelName)
	if err != nil {
		return nil, nil, err
	}

	routingInstances, err := policygroup.BuildRoutingInstancesForClient(ctx, clientID)
	if err != nil {
		return nil, nil, err
	}
	pickedAt := deps.pickedAt
	if pickedAt.IsZero() {
		pickedAt = time.Now()
	}

	attempt := modelPickAttempt{
		clientID:         clientID,
		headers:          headers,
		group:            group,
		routingInstances: routingInstances,
		meta:             policygroup.BuildRequestMetaFromHeader(headers),
		routeEngine:      deps.routeEngine,
		healthManager:    deps.healthManager,
		pickedAt:         pickedAt,
	}

	trace, instance, err := pickModelInstanceFromPolicyGroup(ctx, attempt)
	if err != nil {
		trace, instance, fallbackErr := pickModelInstanceFromUnhealthy(ctx, attempt, err)
		if fallbackErr == nil && instance != nil {
			return trace, instance, nil
		}
		if fallbackErr != nil && !errors.Is(fallbackErr, err) {
			return nil, nil, fallbackErr
		}
		return nil, nil, err
	}
	return trace, instance, nil
}

// pickModelInstanceFromPolicyGroup performs the normal policy-group-based pick.
// This path respects the regular routing rules and request-level retry exclusion.
func pickModelInstanceFromPolicyGroup(
	ctx context.Context,
	attempt modelPickAttempt,
) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	if attempt.routeEngine == nil {
		return nil, nil, fmt.Errorf("nil policy group engine")
	}
	routingInstances := filterRetryExcludedInstances(ctx, attempt.routingInstances)
	instance, trace, err := attempt.routeEngine.Route(ctx, policygroup.RouteRequest{
		ClientID:  attempt.clientID,
		Group:     attempt.group,
		Instances: routingInstances,
		Meta:      attempt.meta,
		Ctx:       ctx,
	})
	if err != nil {
		return nil, nil, err
	}
	return trace, instance, nil
}

func filterRetryExcludedInstances(ctx context.Context, instances []*policygroup.RoutingModelInstance) []*policygroup.RoutingModelInstance {
	excluded, ok := modelretry.GetExcludedModelIDs(ctx)
	if !ok || len(excluded) == 0 {
		return instances
	}
	filtered := make([]*policygroup.RoutingModelInstance, 0, len(instances))
	for _, instance := range instances {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		if _, hit := excluded[instance.ModelWithProvider.Id]; hit {
			continue
		}
		filtered = append(filtered, instance)
	}
	return filtered
}
