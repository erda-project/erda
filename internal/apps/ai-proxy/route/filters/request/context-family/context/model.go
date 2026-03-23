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

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	pgengine "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/engine"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	groupresolver "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/resolver"
	modelretry "github.com/erda-project/erda/internal/apps/ai-proxy/route/reverse_proxy/model_retry"
)

func findModel(req *http.Request, client *clientpb.Client) (*modelpb.Model, error) {
	identifier, err := findModelIdentifier(req)
	if err != nil {
		return nil, fmt.Errorf("failed to find model: %v", err)
	}
	if identifier == "" {
		return nil, fmt.Errorf("missing model")
	}

	ctx := req.Context()
	trace, instance, err := routeToModelInstance(ctx, client.Id, identifier, req.Header)
	if err != nil {
		return nil, err
	}

	// render model by template for downstream usage
	model, err := cachehelpers.GetRenderedModelByID(ctx, instance.ModelWithProvider.Id)
	if err != nil {
		return nil, err
	}

	ctxhelper.PutPolicyTrace(ctx, trace)

	return model, nil
}

func routeToModelInstance(ctx context.Context, clientID, modelName string, headers http.Header) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	return routeToModelInstanceWithDeps(ctx, modelRouteInput{
		clientID:  clientID,
		modelName: modelName,
		headers:   headers,
	}, modelRouteDeps{
		routeEngine:   pgengine.GetEngine(),
		healthManager: health.GetManager(),
		now:           time.Now,
	})
}

type modelRouteInput struct {
	clientID  string
	modelName string
	headers   http.Header
}

type modelRouteContext struct {
	clientID         string
	group            *policypb.PolicyGroup
	routingInstances []*policygroup.RoutingModelInstance
	meta             policygroup.RequestMeta
}

type modelRouteDeps struct {
	routeEngine   *pgengine.Engine
	healthManager *health.Manager
	now           func() time.Time
}

func routeToModelInstanceWithDeps(
	ctx context.Context,
	input modelRouteInput,
	deps modelRouteDeps,
) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	routeCtx, err := resolveModelRouteContext(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	trace, instance, err := routeWithEngine(ctx, routeCtx, deps.routeEngine)
	if err != nil {
		trace, instance, fallbackErr := tryRouteRetryUnhealthy(ctx, routeCtx, err, deps)
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

func resolveModelRouteContext(ctx context.Context, input modelRouteInput) (*modelRouteContext, error) {
	resolver := groupresolver.NewResolver()
	group, err := resolver.Resolve(ctx, input.clientID, input.modelName)
	if err != nil {
		return nil, err
	}

	routingInstances, err := policygroup.BuildRoutingInstancesForClient(ctx, input.clientID)
	if err != nil {
		return nil, err
	}

	return &modelRouteContext{
		clientID:         input.clientID,
		group:            group,
		routingInstances: routingInstances,
		meta:             policygroup.BuildRequestMetaFromHeader(input.headers),
	}, nil
}

func routeWithEngine(
	ctx context.Context,
	routeCtx *modelRouteContext,
	routeEngine *pgengine.Engine,
) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	if routeEngine == nil {
		return nil, nil, fmt.Errorf("nil policy group engine")
	}
	if routeCtx == nil {
		return nil, nil, fmt.Errorf("nil model route context")
	}
	routingInstances := filterRetryExcludedInstances(ctx, routeCtx.routingInstances)
	instance, trace, err := routeEngine.Route(ctx, policygroup.RouteRequest{
		ClientID:  routeCtx.clientID,
		Group:     routeCtx.group,
		Instances: routingInstances,
		Meta:      routeCtx.meta,
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

func coalesceNow(now func() time.Time) func() time.Time {
	if now != nil {
		return now
	}
	return time.Now
}
