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
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

const (
	retryUnhealthyTraceBranchName   = "retry_unhealthy"
	retryUnhealthyTraceStrategyName = "retry_unhealthy"
	retryUnhealthySourceCurrent     = "current-session-unhealthy"
	retryUnhealthySourceOther       = "other-session-unhealthy"
)

type retryUnhealthyCandidate struct {
	instance       *policygroup.RoutingModelInstance
	source         string
	markedAt       time.Time
	currentSession bool
}

func tryRouteRetryUnhealthy(
	ctx context.Context,
	attempt modelRouteAttempt,
	routeErr error,
) (*policygroup.RouteTrace, *policygroup.RoutingModelInstance, error) {
	if !shouldRetryUnhealthy(ctx, routeErr) {
		return nil, nil, routeErr
	}

	candidate, staleFilteredCount, err := pickRetryUnhealthyCandidate(ctx, attempt)
	if err != nil {
		return nil, nil, err
	}
	if candidate == nil {
		return nil, nil, routeErr
	}

	audithelper.Note(ctx, "model_retry.unhealthy_fallback", true)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_instance_id", candidate.instance.ModelWithProvider.Id)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_source", candidate.source)
	audithelper.Note(ctx, "model_retry.unhealthy_fallback_filtered_stale_count", staleFilteredCount)

	return buildRetryUnhealthyTrace(attempt.group), candidate.instance, nil
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

func pickRetryUnhealthyCandidate(
	ctx context.Context,
	attempt modelRouteAttempt,
) (*retryUnhealthyCandidate, int, error) {
	if attempt.healthManager == nil || attempt.group == nil {
		return nil, 0, nil
	}

	window := attempt.healthManager.RetryUnhealthyFallbackWindow()
	if window <= 0 {
		return nil, 0, nil
	}
	cutoff := attempt.now().Add(-window)
	sessionMarks := getRetrySessionUnhealthyMarks(ctx)

	var candidates []retryUnhealthyCandidate
	staleFilteredCount := 0
	for _, instance := range allGroupInstancesForRetryUnhealthy(attempt.group, attempt.routingInstances) {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		state, ok, err := attempt.healthManager.GetState(ctx, attempt.clientID, instance.ModelWithProvider.Id)
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

func getRetrySessionUnhealthyMarks(ctx context.Context) map[string]time.Time {
	raw, ok := ctxhelper.GetModelRetrySessionUnhealthyMarks(ctx)
	if !ok || raw == nil {
		return nil
	}
	marks, _ := raw.(map[string]time.Time)
	return marks
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
