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
	"sort"
	"time"

	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	pgengine "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/engine"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	groupresolver "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/resolver"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

type RetryRouteMode string

const (
	RetryRouteModeNone      RetryRouteMode = "none"
	RetryRouteModeNormal    RetryRouteMode = "normal"
	RetryRouteModeUnhealthy RetryRouteMode = "unhealthy"
)

const (
	retryUnhealthyTraceBranchName   = "retry_unhealthy"
	retryUnhealthySourceCurrent     = "current-session-unhealthy"
	retryUnhealthySourceOther       = "other-session-unhealthy"
	retryUnhealthyTraceStrategyName = "retry_unhealthy"
)

type retrySessionUnhealthyMarks map[string]time.Time

type retryUnhealthyCandidate struct {
	instance       *policygroup.RoutingModelInstance
	source         string
	markedAt       time.Time
	currentSession bool
}

func PredictRetryRouteMode(
	ctx context.Context,
	clientID string,
	healthManager *health.Manager,
	now func() time.Time,
) (RetryRouteMode, error) {
	group, err := resolveRetryRouteGroup(ctx, clientID)
	if err != nil {
		return RetryRouteModeNone, err
	}
	if group == nil {
		return RetryRouteModeNone, nil
	}
	routingInstances, err := policygroup.BuildRoutingInstancesForClient(ctx, clientID)
	if err != nil {
		return RetryRouteModeNone, err
	}

	routeCtx := &modelRouteContext{
		clientID:         clientID,
		group:            group,
		routingInstances: routingInstances,
	}

	hasNormal, err := hasNormalRouteCandidate(ctx, routeCtx, healthManager)
	if err != nil {
		return RetryRouteModeNone, err
	}
	if hasNormal {
		return RetryRouteModeNormal, nil
	}

	candidate, _, err := pickRetryUnhealthyCandidate(ctx, routeCtx, healthManager, coalesceNow(now))
	if err != nil {
		return RetryRouteModeNone, err
	}
	if candidate != nil {
		return RetryRouteModeUnhealthy, nil
	}
	return RetryRouteModeNone, nil
}

func resolveRetryRouteGroup(ctx context.Context, clientID string) (*policypb.PolicyGroup, error) {
	if clientID == "" {
		return nil, nil
	}

	var candidates []string
	if trace := currentPolicyTrace(ctx); trace != nil && trace.Group.Name != "" {
		candidates = append(candidates, trace.Group.Name)
	}
	if model, ok := ctxhelper.GetModel(ctx); ok && model != nil && model.Id != "" {
		if len(candidates) == 0 || candidates[len(candidates)-1] != model.Id {
			candidates = append(candidates, model.Id)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	resolver := groupresolver.NewResolver()
	var lastErr error
	for _, identifier := range candidates {
		group, err := resolver.Resolve(ctx, clientID, identifier)
		if err == nil && group != nil {
			return group, nil
		}
		if err != nil {
			lastErr = err
		}
	}
	return nil, lastErr
}

func currentPolicyTrace(ctx context.Context) *policygroup.RouteTrace {
	traceVal, ok := ctxhelper.GetPolicyTrace(ctx)
	if !ok || traceVal == nil {
		return nil
	}
	trace, _ := traceVal.(*policygroup.RouteTrace)
	return trace
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

	candidate, staleFilteredCount, err := pickRetryUnhealthyCandidate(ctx, routeCtx, deps.healthManager, coalesceNow(deps.now))
	if err != nil {
		return nil, nil, err
	}
	if candidate == nil {
		return nil, nil, routeErr
	}

	count, _ := ctxhelper.GetModelRetryUnhealthyFallbackCount(ctx)
	ctxhelper.PutModelRetryUnhealthyFallbackCount(ctx, count+1)

	audithelper.Note(ctx, "model_retry.unhealthy_fallback", true)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_instance_id", candidate.instance.ModelWithProvider.Id)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_source", candidate.source)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_filtered_stale_count", staleFilteredCount)

	return buildRetryUnhealthyTrace(routeCtx.group), candidate.instance, nil
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

func hasNormalRouteCandidate(
	ctx context.Context,
	routeCtx *modelRouteContext,
	healthManager *health.Manager,
) (bool, error) {
	if routeCtx == nil {
		return false, nil
	}
	filtered := filterRetryExcludedInstances(ctx, routeCtx.routingInstances)
	filtered, err := filterHealthyInstancesReadOnly(ctx, routeCtx.clientID, filtered, healthManager)
	if err != nil {
		return false, err
	}
	return len(allGroupInstancesForRetryUnhealthy(routeCtx.group, filtered)) > 0, nil
}

func pickRetryUnhealthyCandidate(
	ctx context.Context,
	routeCtx *modelRouteContext,
	healthManager *health.Manager,
	now func() time.Time,
) (*retryUnhealthyCandidate, int, error) {
	if healthManager == nil || routeCtx == nil || routeCtx.group == nil {
		return nil, 0, nil
	}

	window := healthManager.RetryUnhealthyFallbackWindow()
	if window <= 0 {
		return nil, 0, nil
	}
	cutoff := coalesceNow(now)().Add(-window)
	sessionMarks := getRetrySessionUnhealthyMarks(ctx)

	var candidates []retryUnhealthyCandidate
	staleFilteredCount := 0
	for _, instance := range allGroupInstancesForRetryUnhealthy(routeCtx.group, routeCtx.routingInstances) {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		state, ok, err := healthManager.GetState(ctx, routeCtx.clientID, instance.ModelWithProvider.Id)
		if err != nil {
			return nil, 0, err
		}
		if !ok || state == nil || state.State != "unhealthy" {
			continue
		}
		if state.UpdatedAt.Before(cutoff) {
			staleFilteredCount++
			continue
		}
		candidate := retryUnhealthyCandidate{
			instance: instance,
			source:   retryUnhealthySourceOther,
			markedAt: state.UpdatedAt,
		}
		if markedAt, ok := sessionMarks[instance.ModelWithProvider.Id]; ok && !markedAt.IsZero() {
			candidate.currentSession = true
			candidate.source = retryUnhealthySourceCurrent
			candidate.markedAt = markedAt
		}
		candidates = append(candidates, candidate)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].currentSession != candidates[j].currentSession {
			return candidates[i].currentSession
		}
		if !candidates[i].markedAt.Equal(candidates[j].markedAt) {
			return candidates[i].markedAt.After(candidates[j].markedAt)
		}
		return candidates[i].instance.ModelWithProvider.Id < candidates[j].instance.ModelWithProvider.Id
	})

	if len(candidates) == 0 {
		return nil, staleFilteredCount, nil
	}
	return &candidates[0], staleFilteredCount, nil
}

func getRetrySessionUnhealthyMarks(ctx context.Context) retrySessionUnhealthyMarks {
	raw, ok := ctxhelper.GetModelRetrySessionUnhealthyMarks(ctx)
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case retrySessionUnhealthyMarks:
		return v
	case map[string]time.Time:
		return retrySessionUnhealthyMarks(v)
	default:
		return nil
	}
}

func allGroupInstancesForRetryUnhealthy(
	group *policypb.PolicyGroup,
	routingInstances []*policygroup.RoutingModelInstance,
) []*policygroup.RoutingModelInstance {
	if group == nil || len(group.Branches) == 0 || len(routingInstances) == 0 {
		return nil
	}
	// unhealthy fallback is the terminal rescue path after normal routing is exhausted.
	// It intentionally scans every instance that belongs to the resolved group, and does
	// not re-apply request-level retry exclusion or normal branch RR/weight semantics.
	seen := make(map[string]struct{})
	ret := make([]*policygroup.RoutingModelInstance, 0, len(routingInstances))
	for _, branch := range group.Branches {
		for _, instance := range selector.MatchSelector(routingInstances, branch.Selector) {
			if instance == nil || instance.ModelWithProvider == nil {
				continue
			}
			id := instance.ModelWithProvider.Id
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ret = append(ret, instance)
		}
	}
	return ret
}

func filterHealthyInstancesReadOnly(
	ctx context.Context,
	clientID string,
	instances []*policygroup.RoutingModelInstance,
	healthManager *health.Manager,
) ([]*policygroup.RoutingModelInstance, error) {
	if healthManager == nil || len(instances) == 0 {
		return instances, nil
	}
	filtered := make([]*policygroup.RoutingModelInstance, 0, len(instances))
	for _, instance := range instances {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		state, ok, err := healthManager.GetState(ctx, clientID, instance.ModelWithProvider.Id)
		if err != nil {
			return nil, err
		}
		if ok && state != nil && state.State == "unhealthy" {
			continue
		}
		filtered = append(filtered, instance)
	}
	return filtered, nil
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
	if remainingAttempts <= 1 {
		return time.Second
	}
	switch {
	case fallbackIndex <= 1:
		return 0
	case fallbackIndex == 2:
		return time.Second
	default:
		return 2 * time.Second
	}
}

func nextRetryUnhealthyDelay(remainingAttempts, fallbackIndex int) time.Duration {
	return NextRetryUnhealthyDelay(remainingAttempts, fallbackIndex)
}
