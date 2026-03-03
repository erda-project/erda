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

package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestFilterHealthyInstances(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     20 * time.Millisecond,
		},
	})

	unhealthy := ModelHealthState{
		State:     stateUnhealthy,
		APIType:   APITypeChatCompletions,
		UpdatedAt: time.Now(),
	}
	data, err := json.Marshal(unhealthy)
	if err != nil {
		t.Fatalf("marshal unhealthy state failed: %v", err)
	}
	if err := store.SetBinding(context.Background(), modelHealthBindingKey, "i1", string(data), time.Hour); err != nil {
		t.Fatalf("set unhealthy state failed: %v", err)
	}

	instances := []*policygroup.RoutingModelInstance{
		testRoutingInstance("i1"),
		testRoutingInstance("i2"),
	}
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	filtered := manager.FilterHealthyInstances(policygroup.RouteRequest{Ctx: ctx}, instances)
	if len(filtered) != 1 || filtered[0].ModelWithProvider.Id != "i2" {
		t.Fatalf("expected only i2 after health filter, got %v", collectIDs(filtered))
	}
	meta, ok := ctxhelper.GetPolicyGroupHealthMeta(ctx)
	if !ok || meta == nil {
		t.Fatal("expected policy group health meta in ctx")
	}
	if len(meta.FilteredUnhealthyInstanceIDs) != 1 || meta.FilteredUnhealthyInstanceIDs[0] != "i1" {
		t.Fatalf("expected filtered unhealthy instance id i1, got %#v", meta.FilteredUnhealthyInstanceIDs)
	}

	// Probe request should bypass health filter so unhealthy instance can be checked.
	probeReq := policygroup.RouteRequest{
		Meta: policygroup.RequestMeta{
			Keys: map[string]string{
				common_types.StickyKeyPrefixFromReqHeader + strings.ToLower(vars.XAIProxyHealthProbe): "true",
			},
		},
	}
	filteredProbe := manager.FilterHealthyInstances(probeReq, instances)
	if len(filteredProbe) != 2 {
		t.Fatalf("expected probe request to bypass health filter, got %v", collectIDs(filteredProbe))
	}
}

func TestNewManagerPanicWhenRescueBackoffUnset(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when rescue backoff is unset")
		}
	}()
	_ = NewManager(store, Config{})
}

func TestMarkUnhealthyStartsSingleProbeWorker(t *testing.T) {
	t.Parallel()

	var probeHits int32
	firstProbeStarted := make(chan struct{}, 1)
	releaseFirstProbe := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&probeHits, 1) == 1 {
			firstProbeStarted <- struct{}{}
			<-releaseFirstProbe
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		ProbeBaseURL: server.URL,
		UnhealthyTTL: time.Hour,
		ProbeTimeout: 2 * time.Second,
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     20 * time.Millisecond,
		},
	})

	headers := http.Header{"Authorization": []string{"Bearer t_test"}}
	manager.MarkUnhealthy(context.Background(), "i1", APITypeChatCompletions, "read tcp: connection reset by peer", headers)

	select {
	case <-firstProbeStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("probe worker did not start")
	}

	// Mark the same instance unhealthy again while first probe is still running.
	manager.MarkUnhealthy(context.Background(), "i1", APITypeChatCompletions, "read tcp: connection reset by peer", headers)
	time.Sleep(100 * time.Millisecond)

	if got := atomic.LoadInt32(&probeHits); got != 1 {
		t.Fatalf("expected single probe worker, got %d probe requests", got)
	}

	close(releaseFirstProbe)

	waitFor(t, 2*time.Second, func() bool {
		_, ok, _ := manager.GetState(context.Background(), "i1")
		return !ok
	})
}

func TestProbeNon2xxShouldKeepUnhealthy(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"context deadline exceeded"}`))
	}))
	defer server.Close()

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		ProbeBaseURL: server.URL,
		UnhealthyTTL: 500 * time.Millisecond,
		ProbeTimeout: 200 * time.Millisecond,
		Rescue: RescueConfig{
			InitialBackoff: 20 * time.Millisecond,
			MaxBackoff:     20 * time.Millisecond,
		},
	})

	manager.MarkUnhealthy(context.Background(), "i-non2xx", APITypeChatCompletions, "network timeout", http.Header{
		vars.XAIProxyGeneratedCallId: []string{"call-1"},
	})

	// Give probe worker enough time to attempt and fail at least once.
	time.Sleep(150 * time.Millisecond)

	state, ok, err := manager.GetState(context.Background(), "i-non2xx")
	if err != nil {
		t.Fatalf("get state failed: %v", err)
	}
	if !ok || state == nil {
		t.Fatal("expected unhealthy state to remain after non-2xx probe response")
	}
	if state.State != stateUnhealthy {
		t.Fatalf("expected state=%s, got %s", stateUnhealthy, state.State)
	}
}

func TestUnhealthyTTLRefreshedOnProbeFailure(t *testing.T) {
	t.Parallel()

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		ProbeBaseURL: "http://127.0.0.1:1",
		UnhealthyTTL: 40 * time.Millisecond,
		ProbeTimeout: 20 * time.Millisecond,
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	})

	manager.MarkUnhealthy(context.Background(), "i-fail", APITypeChatCompletions, "connection reset by peer", http.Header{
		"Authorization": []string{"Bearer t_fail"},
	})

	// If unhealthy TTL is not refreshed by failed probes, the state should expire quickly.
	// We wait much longer than unhealthy_ttl and still expect unhealthy state to exist.
	time.Sleep(220 * time.Millisecond)

	state, ok, err := manager.GetState(context.Background(), "i-fail")
	if err != nil {
		t.Fatalf("get state failed: %v", err)
	}
	if !ok || state == nil {
		t.Fatal("expected unhealthy state kept by failed probe refresh, but state expired")
	}
	if state.State != stateUnhealthy {
		t.Fatalf("expected state=%s, got %s", stateUnhealthy, state.State)
	}
}

func TestBuildProbeHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer t_x")
	headers.Set("AK-Token", "ak-token-x")
	headers.Set("X-Trace-Id", "trace-1")
	headers.Set(vars.XAIProxyHealthProbe, "true")
	probeHeaders := BuildProbeHeaders(headers)

	if probeHeaders.Get("Authorization") == "" || probeHeaders.Get("AK-Token") == "" {
		t.Fatalf("expected headers kept, got: %v", probeHeaders)
	}
	if probeHeaders.Get("X-Trace-Id") != "trace-1" {
		t.Fatalf("expected x-trace-id kept, got: %v", probeHeaders)
	}
	if probeHeaders.Get(vars.XAIProxyHealthProbe) != "true" {
		t.Fatalf("expected probe marker kept, got: %v", probeHeaders)
	}
}

func TestFilterUnhealthyRearmWorkerAfterRestart(t *testing.T) {
	t.Parallel()

	var probeHits int32
	var gotAuth atomic.Value
	var gotTraceID atomic.Value
	gotAuth.Store("")
	gotTraceID.Store("")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth.Store(r.Header.Get("Authorization"))
		gotTraceID.Store(r.Header.Get("X-Trace-Id"))
		atomic.AddInt32(&probeHits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		ProbeBaseURL: server.URL,
		UnhealthyTTL: time.Hour,
		ProbeTimeout: 2 * time.Second,
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     20 * time.Millisecond,
		},
	})

	stateBytes, err := json.Marshal(&ModelHealthState{
		State:     stateUnhealthy,
		APIType:   APITypeChatCompletions,
		LastError: "read tcp: connection reset by peer",
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("marshal unhealthy state failed: %v", err)
	}
	if err := store.SetBinding(context.Background(), modelHealthBindingKey, "i1", string(stateBytes), time.Hour); err != nil {
		t.Fatalf("set unhealthy state failed: %v", err)
	}

	filtered := manager.FilterHealthyInstances(policygroup.RouteRequest{
		Meta: policygroup.RequestMeta{
			Keys: map[string]string{
				common_types.StickyKeyPrefixFromReqHeader + strings.ToLower("Authorization"): "Bearer t_from_request",
				common_types.StickyKeyPrefixFromReqHeader + strings.ToLower("X-Trace-Id"):    "trace-123",
			},
		},
	}, []*policygroup.RoutingModelInstance{
		testRoutingInstance("i1"),
		testRoutingInstance("i2"),
	})
	if len(filtered) != 1 || filtered[0].ModelWithProvider.Id != "i2" {
		t.Fatalf("expected i1 filtered out and i2 kept, got %v", collectIDs(filtered))
	}

	waitFor(t, 2*time.Second, func() bool {
		_, ok, _ := manager.GetState(context.Background(), "i1")
		return !ok
	})
	if atomic.LoadInt32(&probeHits) == 0 {
		t.Fatal("expected re-armed worker to trigger probe request")
	}
	if gotAuth.Load().(string) != "Bearer t_from_request" {
		t.Fatalf("expected authorization forwarded from request meta, got %q", gotAuth.Load().(string))
	}
	if gotTraceID.Load().(string) != "trace-123" {
		t.Fatalf("expected x-trace-id forwarded from request meta, got %q", gotTraceID.Load().(string))
	}
}

func TestUnsupportedAPITypeDoesNotKeepUnhealthyForever(t *testing.T) {
	t.Parallel()

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		ProbeBaseURL: "http://127.0.0.1:1",
		UnhealthyTTL: time.Hour,
		ProbeTimeout: 20 * time.Millisecond,
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	})

	manager.MarkUnhealthy(context.Background(), "i-unsupported", APIType("unsupported"), "network error", http.Header{
		"Authorization": []string{"Bearer t_unsupported"},
	})

	waitFor(t, 2*time.Second, func() bool {
		_, ok, _ := manager.GetState(context.Background(), "i-unsupported")
		return !ok
	})
}

func TestMarkUnhealthyWritesModelMarkHeaderContext(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		ProbeBaseURL: "http://127.0.0.1:1",
		UnhealthyTTL: time.Hour,
		ProbeTimeout: 20 * time.Millisecond,
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	})

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	manager.MarkUnhealthy(ctx, "m-ctx", APITypeChatCompletions, "network timeout", http.Header{})
	got, ok := ctxhelper.GetModelMarkUnhealthyInstanceID(ctx)
	if !ok || got != "m-ctx" {
		t.Fatalf("expected model mark unhealthy instance id m-ctx, got %q", got)
	}
}

func TestFilterUnhealthyUnsupportedAPITypeReleasedAndRecorded(t *testing.T) {
	t.Parallel()

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		ProbeBaseURL: "http://127.0.0.1:1",
		UnhealthyTTL: time.Hour,
		ProbeTimeout: 20 * time.Millisecond,
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	})

	stateBytes, err := json.Marshal(&ModelHealthState{
		State:     stateUnhealthy,
		APIType:   APIType("embeddings"),
		LastError: "read tcp: connection reset by peer",
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("marshal unhealthy state failed: %v", err)
	}
	if err := store.SetBinding(context.Background(), modelHealthBindingKey, "i1", string(stateBytes), time.Hour); err != nil {
		t.Fatalf("set unhealthy state failed: %v", err)
	}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	filtered := manager.FilterHealthyInstances(policygroup.RouteRequest{
		Ctx: ctx,
	}, []*policygroup.RoutingModelInstance{
		testRoutingInstance("i1"),
		testRoutingInstance("i2"),
	})
	if len(filtered) != 2 {
		t.Fatalf("expected unsupported api_type instance released, got %v", collectIDs(filtered))
	}
	if _, ok, _ := manager.GetState(context.Background(), "i1"); ok {
		t.Fatal("expected unhealthy state deleted for unsupported api_type")
	}
	meta, ok := ctxhelper.GetPolicyGroupHealthMeta(ctx)
	if !ok || meta == nil {
		t.Fatal("expected policy group health meta in ctx")
	}
	if meta.ReleasedUnsupportedCount != 1 {
		t.Fatalf("expected released unsupported count=1, got %d", meta.ReleasedUnsupportedCount)
	}
	if len(meta.ReleasedUnsupportedAPITypes) != 1 || meta.ReleasedUnsupportedAPITypes[0] != "embeddings" {
		t.Fatalf("expected released unsupported api type embeddings, got %#v", meta.ReleasedUnsupportedAPITypes)
	}
	if len(meta.FilteredUnhealthyInstanceIDs) != 0 {
		t.Fatalf("expected no filtered unhealthy instance ids, got %#v", meta.FilteredUnhealthyInstanceIDs)
	}
}

func testRoutingInstance(id string) *policygroup.RoutingModelInstance {
	return &policygroup.RoutingModelInstance{
		ModelWithProvider: &cachehelpers.ModelWithProvider{
			Model: &modelpb.Model{
				Id:   id,
				Name: id,
			},
		},
	}
}

func collectIDs(instances []*policygroup.RoutingModelInstance) []string {
	ids := make([]string, 0, len(instances))
	for _, instance := range instances {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		ids = append(ids, instance.ModelWithProvider.Id)
	}
	return ids
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}
