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
	"time"

	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	pgengine "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/engine"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	retryunhealthy "github.com/erda-project/erda/internal/apps/ai-proxy/route/retry_unhealthy"
)

type RetryRouteMode = retryunhealthy.RouteMode

const (
	RetryRouteModeNone      = retryunhealthy.RouteModeNone
	RetryRouteModeNormal    = retryunhealthy.RouteModeNormal
	RetryRouteModeUnhealthy = retryunhealthy.RouteModeUnhealthy
)

const (
	retryUnhealthyTraceBranchName   = "retry_unhealthy"
	retryUnhealthyTraceStrategyName = "retry_unhealthy"
)

func PredictRetryRouteMode(
	ctx context.Context,
	clientID string,
	healthManager *health.Manager,
	now func() time.Time,
) (RetryRouteMode, error) {
	return retryunhealthy.PredictRouteMode(ctx, clientID, healthManager, now)
}

func tryRouteRetryUnhealthy(
	ctx context.Context,
	routeCtx *modelRouteContext,
	routeErr error,
	deps modelRouteDeps,
) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	if !shouldRetryUnhealthy(ctx, routeErr) {
		return nil, nil, routeErr
	}
	if routeCtx == nil {
		return nil, nil, routeErr
	}

	candidate, staleFilteredCount, err := retryunhealthy.PickCandidate(ctx, retryunhealthy.RouteContext{
		ClientID:         routeCtx.clientID,
		Group:            routeCtx.group,
		RoutingInstances: routeCtx.routingInstances,
	}, deps.healthManager, deps.now)
	if err != nil {
		return nil, nil, err
	}
	if candidate == nil {
		return nil, nil, routeErr
	}

	count, _ := ctxhelper.GetModelRetryUnhealthyFallbackCount(ctx)
	ctxhelper.PutModelRetryUnhealthyFallbackCount(ctx, count+1)

	audithelper.Note(ctx, "model_retry.unhealthy_fallback", true)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_instance_id", candidate.Instance.ModelWithProvider.Id)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_source", candidate.Source)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_filtered_stale_count", staleFilteredCount)

	return buildRetryUnhealthyTrace(routeCtx.group), candidate.Instance, nil
}

func shouldRetryUnhealthy(ctx context.Context, routeErr error) bool {
	if routeErr == nil {
		return false
	}
	attempt, ok := ctxhelper.GetModelRetryRawLLMBackendRequestCount(ctx)
	if !ok || attempt <= 1 {
		return false
	}
	return errors.Is(routeErr, pgengine.ErrNoAvailableBranch) || errors.Is(routeErr, pgengine.ErrNoAvailableInstance)
}

func allGroupInstancesForRetryUnhealthy(
	group *policypb.PolicyGroup,
	routingInstances []*policygroup.RoutingModelInstance,
) []*policygroup.RoutingModelInstance {
	// unhealthy fallback is the terminal rescue path after normal routing is exhausted.
	// It intentionally scans every instance that belongs to the resolved group, and does
	// not re-apply request-level retry exclusion or normal branch RR/weight semantics.
	return retryunhealthy.AllGroupInstances(group, routingInstances)
}

func buildRetryUnhealthyTrace(group *policypb.PolicyGroup) *policygroup.RouteTrace {
	if group == nil {
		return nil
	}
	return &policygroup.RouteTrace{
		Group: policygroup.RouteTraceGroup{
			Source: group.Source,
			Name:   group.Name,
			Mode:   group.Mode,
			Desc:   group.Desc,
		},
		Branch: policygroup.RouteTraceBranch{
			Name:     retryUnhealthyTraceBranchName,
			Strategy: retryUnhealthyTraceStrategyName,
		},
	}
}

func NextRetryUnhealthyDelay(remainingAttempts, fallbackIndex int) time.Duration {
	return retryunhealthy.NextDelay(remainingAttempts, fallbackIndex)
}

func nextRetryUnhealthyDelay(remainingAttempts, fallbackIndex int) time.Duration {
	return NextRetryUnhealthyDelay(remainingAttempts, fallbackIndex)
}
