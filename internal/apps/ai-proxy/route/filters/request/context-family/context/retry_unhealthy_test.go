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
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	clientmodelrelationpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachetypes"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/engine"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/health"
)

type retryUnhealthyFakeCacheManager struct {
	models       []*modelpb.Model
	providers    map[string]*providerpb.ServiceProvider
	policyGroups []*policypb.PolicyGroup
}

func (m *retryUnhealthyFakeCacheManager) ListAll(_ context.Context, itemType cachetypes.ItemType) (uint64, any, error) {
	switch itemType {
	case cachetypes.ItemTypeModel:
		return uint64(len(m.models)), m.models, nil
	case cachetypes.ItemTypeClientModelRelation:
		return 0, []*clientmodelrelationpb.ClientModelRelation{}, nil
	case cachetypes.ItemTypePolicyGroup:
		return uint64(len(m.policyGroups)), m.policyGroups, nil
	default:
		return 0, nil, nil
	}
}

func (m *retryUnhealthyFakeCacheManager) GetByID(_ context.Context, itemType cachetypes.ItemType, id string) (any, error) {
	switch itemType {
	case cachetypes.ItemTypeProvider:
		return m.providers[id], nil
	case cachetypes.ItemTypeModel:
		for _, model := range m.models {
			if model.Id == id {
				return model, nil
			}
		}
	}
	return nil, nil
}

func (m *retryUnhealthyFakeCacheManager) TriggerRefresh(_ context.Context, _ ...cachetypes.ItemType) {
}

func TestRouteToModelInstanceWithDeps_NormalRouteWins(t *testing.T) {
	env := newRetryUnhealthyEnv(t)
	now := time.Now()
	env.writeHealthState("m-unhealthy", now)

	trace, instance, err := routeToModelInstanceWithDeps(env.ctx, env.clientID, "gpt-4.1", http.Header{}, env.routeEngine, env.healthManager, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatalf("routeToModelInstanceWithDeps error: %v", err)
	}
	if instance.ModelWithProvider.Id != "m-healthy" {
		t.Fatalf("expected healthy instance selected, got %s", instance.ModelWithProvider.Id)
	}
	if trace == nil || trace.Branch.Name != "auto" {
		t.Fatalf("expected normal trace branch auto, got %#v", trace)
	}
}

func TestRouteToModelInstanceWithDeps_NoFallbackOnFirstAttempt(t *testing.T) {
	env := newRetryUnhealthyEnv(t)
	now := time.Now()
	env.writeHealthState("m-healthy", now)
	env.writeHealthState("m-unhealthy", now)

	_, _, err := routeToModelInstanceWithDeps(env.ctx, env.clientID, "gpt-4.1", http.Header{}, env.routeEngine, env.healthManager, func() time.Time {
		return now
	})
	if err == nil {
		t.Fatal("expected no available route on first attempt")
	}
}

func TestRouteToModelInstanceWithDeps_FallbackPrefersCurrentSessionUnhealthy(t *testing.T) {
	env := newRetryUnhealthyEnv(t)
	now := time.Now()
	ctxhelper.PutModelRetryRawLLMBackendRequestCount(env.ctx, 2)
	ctxhelper.PutModelRetryUnhealthyFallbackCount(env.ctx, 0)
	ctxhelper.PutModelRetrySessionUnhealthyMarks(env.ctx, ctxhelper.ModelRetrySessionUnhealthyMarks{
		"m-healthy":   now.Add(-2 * time.Second),
		"m-unhealthy": now.Add(-1 * time.Second),
	})
	env.writeHealthState("m-healthy", now.Add(-5*time.Second))
	env.writeHealthState("m-unhealthy", now.Add(-30*time.Second))

	trace, instance, err := routeToModelInstanceWithDeps(env.ctx, env.clientID, "gpt-4.1", http.Header{}, env.routeEngine, env.healthManager, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatalf("routeToModelInstanceWithDeps error: %v", err)
	}
	if instance.ModelWithProvider.Id != "m-unhealthy" {
		t.Fatalf("expected current-session unhealthy instance selected, got %s", instance.ModelWithProvider.Id)
	}
	if trace == nil || trace.Branch.Name != retryUnhealthyTraceBranchName {
		t.Fatalf("expected retry unhealthy trace, got %#v", trace)
	}
}

func TestRouteToModelInstanceWithDeps_FallbackPrefersRecentOtherSessionUnhealthy(t *testing.T) {
	env := newRetryUnhealthyEnv(t)
	now := time.Now()
	ctxhelper.PutModelRetryRawLLMBackendRequestCount(env.ctx, 2)
	env.writeHealthState("m-healthy", now.Add(-8*time.Minute))
	env.writeHealthState("m-unhealthy", now.Add(-30*time.Second))

	_, instance, err := routeToModelInstanceWithDeps(env.ctx, env.clientID, "gpt-4.1", http.Header{}, env.routeEngine, env.healthManager, func() time.Time {
		return now
	})
	if err != nil {
		t.Fatalf("routeToModelInstanceWithDeps error: %v", err)
	}
	if instance.ModelWithProvider.Id != "m-unhealthy" {
		t.Fatalf("expected recent unhealthy instance selected, got %s", instance.ModelWithProvider.Id)
	}
}

func TestRouteToModelInstanceWithDeps_SkipsStaleUnhealthy(t *testing.T) {
	env := newRetryUnhealthyEnv(t)
	now := time.Now()
	ctxhelper.PutModelRetryRawLLMBackendRequestCount(env.ctx, 2)
	env.writeHealthState("m-healthy", now.Add(-11*time.Minute))
	env.writeHealthState("m-unhealthy", now.Add(-12*time.Minute))

	_, _, err := routeToModelInstanceWithDeps(env.ctx, env.clientID, "gpt-4.1", http.Header{}, env.routeEngine, env.healthManager, func() time.Time {
		return now
	})
	if err == nil {
		t.Fatal("expected stale unhealthy instances skipped")
	}
}

func TestNextRetryUnhealthyDelay(t *testing.T) {
	cases := []struct {
		name              string
		remainingAttempts int
		fallbackIndex     int
		want              time.Duration
	}{
		{name: "first fallback immediate when more chances remain", remainingAttempts: 2, fallbackIndex: 1, want: 0},
		{name: "second fallback waits one second", remainingAttempts: 3, fallbackIndex: 2, want: time.Second},
		{name: "third fallback waits two seconds", remainingAttempts: 4, fallbackIndex: 3, want: 2 * time.Second},
		{name: "last chance waits one second", remainingAttempts: 1, fallbackIndex: 1, want: time.Second},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextRetryUnhealthyDelay(tt.remainingAttempts, tt.fallbackIndex); got != tt.want {
				t.Fatalf("nextRetryUnhealthyDelay(%d, %d)=%s, want=%s", tt.remainingAttempts, tt.fallbackIndex, got, tt.want)
			}
		})
	}
}

func TestRetryUnhealthyMarksCarryAcrossResetForRetry(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	now := time.Now()
	ctxhelper.PutModelRetrySessionUnhealthyMarks(ctx, ctxhelper.ModelRetrySessionUnhealthyMarks{"m-1": now})
	ctxhelper.PutModelRetryUnhealthyFallbackCount(ctx, 2)

	next := ctxhelper.ResetForRetry(ctx)

	marks, ok := ctxhelper.GetModelRetrySessionUnhealthyMarks(next)
	if !ok || len(marks) != 1 || !marks["m-1"].Equal(now) {
		t.Fatalf("expected retry unhealthy marks carried across reset, got %#v", marks)
	}
	count, ok := ctxhelper.GetModelRetryUnhealthyFallbackCount(next)
	if !ok || count != 2 {
		t.Fatalf("expected fallback count carried across reset, got %d", count)
	}
}

type retryUnhealthyEnv struct {
	ctx           context.Context
	clientID      string
	store         state_store.LBStateStore
	healthManager *health.Manager
	routeEngine   *engine.Engine
}

func newRetryUnhealthyEnv(t *testing.T) *retryUnhealthyEnv {
	t.Helper()

	clientID := "client-a"
	providers := map[string]*providerpb.ServiceProvider{
		"p-1": {Id: "p-1", Type: "openai", Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
		"p-2": {Id: "p-2", Type: "azure", Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
	}
	models := []*modelpb.Model{
		{Id: "m-healthy", Name: "gpt-4.1", TemplateId: "gpt-4.1", Publisher: "openai", ClientId: clientID, ProviderId: "p-1", IsEnabled: boolPtr(true), Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
		{Id: "m-unhealthy", Name: "gpt-4.1", TemplateId: "gpt-4.1", Publisher: "openai", ClientId: clientID, ProviderId: "p-2", IsEnabled: boolPtr(true), Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
	}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutCacheManager(ctx, &retryUnhealthyFakeCacheManager{
		models:    models,
		providers: providers,
	})
	ctxhelper.PutClientId(ctx, clientID)

	store := state_store.NewMemoryStateStore()
	manager := health.NewManager(store, health.Config{
		Enabled: true,
		Probe: health.ProbeConfig{
			BaseURL:      "http://127.0.0.1:65530",
			Timeout:      10 * time.Millisecond,
			UnhealthyTTL: time.Hour,
		},
		Rescue: health.RescueConfig{
			InitialBackoff: time.Second,
			MaxBackoff:     2 * time.Minute,
		},
	})
	routeEngine := engine.NewEngine(store, engine.WithHealthFilter(manager.FilterHealthyInstances))

	return &retryUnhealthyEnv{
		ctx:           ctx,
		clientID:      clientID,
		store:         store,
		healthManager: manager,
		routeEngine:   routeEngine,
	}
}

func (e *retryUnhealthyEnv) writeHealthState(instanceID string, updatedAt time.Time) {
	payload, _ := json.Marshal(&health.ModelHealthState{
		State:     "unhealthy",
		APIType:   health.APITypeChatCompletions,
		LastError: "network timeout",
		UpdatedAt: updatedAt,
	})
	_ = e.store.SetBinding(context.Background(), "client:model-health", "client:"+e.clientID+"|instance:"+instanceID, string(payload), time.Hour)
}

func boolPtr(v bool) *bool {
	return &v
}

func TestPredictRetryRouteMode(t *testing.T) {
	env := newRetryUnhealthyEnv(t)
	now := time.Now()
	ctxhelper.PutModelRetryRawLLMBackendRequestCount(env.ctx, 2)
	env.writeHealthState("m-healthy", now)
	env.writeHealthState("m-unhealthy", now)

	req, _ := http.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AI-Proxy-Model-Name", "gpt-4.1")
	req = req.WithContext(env.ctx)
	ctxhelper.PutReverseProxyRequestInSnapshot(env.ctx, req)

	mode, err := PredictRetryRouteMode(env.ctx, req, env.clientID, env.healthManager, func() time.Time { return now })
	if err != nil {
		t.Fatalf("PredictRetryRouteMode error: %v", err)
	}
	if mode != RetryRouteModeUnhealthy {
		t.Fatalf("expected unhealthy retry mode, got %s", mode)
	}
}

func TestAllGroupInstancesForRetryUnhealthyIgnoresBranchRules(t *testing.T) {
	clientID := "client-b"
	providers := map[string]*providerpb.ServiceProvider{
		"p-1": {Id: "p-1", Type: "openai", Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
		"p-2": {Id: "p-2", Type: "azure", Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
	}
	models := []*modelpb.Model{
		{Id: "m-a", Name: "m-a", TemplateId: "gpt-4.1", Publisher: "openai", ClientId: clientID, ProviderId: "p-1", IsEnabled: boolPtr(true), Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
		{Id: "m-b", Name: "m-b", TemplateId: "gpt-4.1", Publisher: "openai", ClientId: clientID, ProviderId: "p-2", IsEnabled: boolPtr(true), Metadata: &metadatapb.Metadata{Public: map[string]*structpb.Value{}}},
	}
	group := &policypb.PolicyGroup{
		ClientId: clientID,
		Name:     "grp",
		Mode:     common_types.PolicyGroupModePriority.String(),
		Branches: []*policypb.PolicyBranch{
			{
				Name:     "b1",
				Priority: 1,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: selectorForInstanceID("m-a"),
			},
			{
				Name:     "b2",
				Priority: 2,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: selectorForInstanceID("m-b"),
			},
		},
	}
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutCacheManager(ctx, &retryUnhealthyFakeCacheManager{
		models:       models,
		providers:    providers,
		policyGroups: []*policypb.PolicyGroup{group},
	})

	instances, err := policygroup.BuildRoutingInstancesForClient(ctx, clientID)
	if err != nil {
		t.Fatalf("BuildRoutingInstancesForClient error: %v", err)
	}
	got := allGroupInstancesForRetryUnhealthy(group, instances)
	if len(got) != 2 {
		t.Fatalf("expected all group instances collected, got %d", len(got))
	}
}

func selectorForInstanceID(instanceID string) *policypb.PolicySelector {
	return &policypb.PolicySelector{
		Requirements: []*policypb.PolicyRequirement{
			{
				Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
				Label: &policypb.LabelRequirement{
					Key:      common_types.PolicyLabelKeyModelInstanceID,
					Operator: common_types.PolicySelectorLabelOpIn.String(),
					Values:   []string{instanceID},
				},
			},
		},
	}
}
