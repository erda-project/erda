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

package retry_unhealthy

import (
	"context"
	"reflect"
	"sort"
	"time"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
	groupresolver "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/resolver"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

type RouteMode string

const (
	RouteModeNone      RouteMode = "none"
	RouteModeNormal    RouteMode = "normal"
	RouteModeUnhealthy RouteMode = "unhealthy"
)

const (
	SourceCurrentSessionUnhealthy = "current-session-unhealthy"
	SourceOtherSessionUnhealthy   = "other-session-unhealthy"
)

type RouteContext struct {
	ClientID         string
	Group            *policypb.PolicyGroup
	RoutingInstances []*policygroup.RoutingModelInstance
}

type Candidate struct {
	Instance       *policygroup.RoutingModelInstance
	Source         string
	MarkedAt       time.Time
	CurrentSession bool
}

func PredictRouteMode(
	ctx context.Context,
	clientID string,
	healthManager *health.Manager,
	now func() time.Time,
) (RouteMode, error) {
	routeCtx, err := resolveRouteContext(ctx, clientID)
	if err != nil {
		return RouteModeNone, err
	}
	if routeCtx == nil {
		return RouteModeNone, nil
	}

	hasNormal, err := HasNormalCandidate(ctx, *routeCtx, healthManager)
	if err != nil {
		return RouteModeNone, err
	}
	if hasNormal {
		return RouteModeNormal, nil
	}

	candidate, _, err := PickCandidate(ctx, *routeCtx, healthManager, now)
	if err != nil {
		return RouteModeNone, err
	}
	if candidate != nil {
		return RouteModeUnhealthy, nil
	}
	return RouteModeNone, nil
}

func HasNormalCandidate(
	ctx context.Context,
	routeCtx RouteContext,
	healthManager *health.Manager,
) (bool, error) {
	filtered := filterRetryExcludedInstances(ctx, routeCtx.RoutingInstances)
	filtered, err := filterHealthyInstancesReadOnly(ctx, routeCtx.ClientID, filtered, healthManager)
	if err != nil {
		return false, err
	}
	return len(AllGroupInstances(routeCtx.Group, filtered)) > 0, nil
}

func PickCandidate(
	ctx context.Context,
	routeCtx RouteContext,
	healthManager *health.Manager,
	now func() time.Time,
) (*Candidate, int, error) {
	if healthManager == nil || routeCtx.Group == nil {
		return nil, 0, nil
	}

	window := healthManager.RetryUnhealthyFallbackWindow()
	if window <= 0 {
		return nil, 0, nil
	}
	cutoff := coalesceNow(now)().Add(-window)
	sessionMarks := getSessionUnhealthyMarks(ctx)

	var candidates []Candidate
	staleFilteredCount := 0
	for _, instance := range AllGroupInstances(routeCtx.Group, routeCtx.RoutingInstances) {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		state, ok, err := healthManager.GetState(ctx, routeCtx.ClientID, instance.ModelWithProvider.Id)
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
		candidate := Candidate{
			Instance: instance,
			Source:   SourceOtherSessionUnhealthy,
			MarkedAt: state.UpdatedAt,
		}
		if markedAt, ok := sessionMarks[instance.ModelWithProvider.Id]; ok && !markedAt.IsZero() {
			candidate.CurrentSession = true
			candidate.Source = SourceCurrentSessionUnhealthy
			candidate.MarkedAt = markedAt
		}
		candidates = append(candidates, candidate)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].CurrentSession != candidates[j].CurrentSession {
			return candidates[i].CurrentSession
		}
		if !candidates[i].MarkedAt.Equal(candidates[j].MarkedAt) {
			return candidates[i].MarkedAt.After(candidates[j].MarkedAt)
		}
		return candidates[i].Instance.ModelWithProvider.Id < candidates[j].Instance.ModelWithProvider.Id
	})

	if len(candidates) == 0 {
		return nil, staleFilteredCount, nil
	}
	return &candidates[0], staleFilteredCount, nil
}

func AllGroupInstances(
	group *policypb.PolicyGroup,
	routingInstances []*policygroup.RoutingModelInstance,
) []*policygroup.RoutingModelInstance {
	if group == nil || len(group.Branches) == 0 || len(routingInstances) == 0 {
		return nil
	}
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

func NextDelay(remainingAttempts, fallbackIndex int) time.Duration {
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

func resolveRouteContext(ctx context.Context, clientID string) (*RouteContext, error) {
	group, err := resolveRouteGroup(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	routingInstances, err := policygroup.BuildRoutingInstancesForClient(ctx, clientID)
	if err != nil {
		return nil, err
	}
	return &RouteContext{
		ClientID:         clientID,
		Group:            group,
		RoutingInstances: routingInstances,
	}, nil
}

func resolveRouteGroup(ctx context.Context, clientID string) (*policypb.PolicyGroup, error) {
	if clientID == "" {
		return nil, nil
	}

	var candidates []string
	if trace := currentPolicyTrace(ctx); trace != nil && trace.Group.Name != "" {
		candidates = append(candidates, trace.Group.Name)
	}
	if model, ok := currentModel(ctx); ok && model != nil && model.Id != "" {
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

func currentModel(ctx context.Context) (*modelpb.Model, bool) {
	model, ok := ctxhelper.GetModel(ctx)
	if !ok || model == nil {
		return nil, false
	}
	return model, true
}

func filterRetryExcludedInstances(
	ctx context.Context,
	instances []*policygroup.RoutingModelInstance,
) []*policygroup.RoutingModelInstance {
	excluded, ok := getExcludedModelIDs(ctx)
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

func getExcludedModelIDs(ctx context.Context) (map[string]struct{}, bool) {
	raw, ok := ctxhelper.GetModelRetryExcludedModelIDs(ctx)
	if !ok || raw == nil {
		return nil, false
	}
	excluded, ok := raw.(map[string]struct{})
	if !ok || excluded == nil {
		return nil, false
	}
	return excluded, true
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

func getSessionUnhealthyMarks(ctx context.Context) map[string]time.Time {
	raw, ok := ctxhelper.GetModelRetrySessionUnhealthyMarks(ctx)
	if !ok || raw == nil {
		return nil
	}
	if marks, ok := raw.(map[string]time.Time); ok {
		return marks
	}

	value := reflect.ValueOf(raw)
	if value.Kind() != reflect.Map || value.Type().Key().Kind() != reflect.String || value.Type().Elem() != reflect.TypeOf(time.Time{}) {
		return nil
	}
	ret := make(map[string]time.Time, value.Len())
	iter := value.MapRange()
	for iter.Next() {
		ret[iter.Key().String()] = iter.Value().Interface().(time.Time)
	}
	return ret
}

func coalesceNow(now func() time.Time) func() time.Time {
	if now != nil {
		return now
	}
	return time.Now
}
