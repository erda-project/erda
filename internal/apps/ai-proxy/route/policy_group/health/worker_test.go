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

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestMarkUnhealthyStartsSingleProbeWorker(t *testing.T) {
	t.Parallel()
	clientID := "client-a"

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
	manager := newTestManager(store, server.URL)

	headers := http.Header{"Authorization": []string{"Bearer t_test"}}
	manager.MarkUnhealthy(context.Background(), clientID, "i1", APITypeChatCompletions, "read tcp: connection reset by peer", headers)

	select {
	case <-firstProbeStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("probe worker did not start")
	}

	manager.MarkUnhealthy(context.Background(), clientID, "i1", APITypeChatCompletions, "read tcp: connection reset by peer", headers)
	time.Sleep(100 * time.Millisecond)

	if got := atomic.LoadInt32(&probeHits); got != 1 {
		t.Fatalf("expected single probe worker, got %d probe requests", got)
	}

	close(releaseFirstProbe)

	waitFor(t, 2*time.Second, func() bool {
		_, ok, _ := manager.GetState(context.Background(), clientID, "i1")
		return !ok
	})
}

func TestProbeNon2xxShouldKeepUnhealthy(t *testing.T) {
	t.Parallel()
	clientID := "client-a"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"context deadline exceeded"}`))
	}))
	defer server.Close()

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		Enabled: true,
		Probe: ProbeConfig{
			BaseURL:      server.URL,
			UnhealthyTTL: 500 * time.Millisecond,
			Timeout:      200 * time.Millisecond,
		},
		Rescue: RescueConfig{
			InitialBackoff: 20 * time.Millisecond,
			MaxBackoff:     20 * time.Millisecond,
		},
	})

	manager.MarkUnhealthy(context.Background(), clientID, "i-non2xx", APITypeChatCompletions, "network timeout", http.Header{
		vars.XAIProxyGeneratedCallId: []string{"call-1"},
	})

	time.Sleep(150 * time.Millisecond)

	state, ok, err := manager.GetState(context.Background(), clientID, "i-non2xx")
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
	clientID := "client-a"

	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{
		Enabled: true,
		Probe: ProbeConfig{
			BaseURL:      "http://127.0.0.1:1",
			UnhealthyTTL: 40 * time.Millisecond,
			Timeout:      20 * time.Millisecond,
		},
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     10 * time.Millisecond,
		},
	})

	manager.MarkUnhealthy(context.Background(), clientID, "i-fail", APITypeChatCompletions, "connection reset by peer", http.Header{
		"Authorization": []string{"Bearer t_fail"},
	})

	time.Sleep(220 * time.Millisecond)

	state, ok, err := manager.GetState(context.Background(), clientID, "i-fail")
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

func TestFilterUnhealthyRearmWorkerAfterRestart(t *testing.T) {
	t.Parallel()
	clientID := "client-a"

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
	manager := newTestManager(store, server.URL)

	stateBytes, err := json.Marshal(&ModelHealthState{
		State:     stateUnhealthy,
		APIType:   APITypeChatCompletions,
		LastError: "read tcp: connection reset by peer",
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("marshal unhealthy state failed: %v", err)
	}
	if err := store.SetBinding(context.Background(), modelHealthBindingKey, makeModelHealthBindingID(clientID, "i1"), string(stateBytes), time.Hour); err != nil {
		t.Fatalf("set unhealthy state failed: %v", err)
	}

	filtered := manager.FilterHealthyInstances(policygroup.RouteRequest{
		ClientID: clientID,
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
		_, ok, _ := manager.GetState(context.Background(), clientID, "i1")
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
	clientID := "client-a"

	store := state_store.NewMemoryStateStore()
	manager := newTestManager(store, "http://127.0.0.1:1")

	manager.MarkUnhealthy(context.Background(), clientID, "i-unsupported", APIType("unsupported"), "network error", http.Header{
		"Authorization": []string{"Bearer t_unsupported"},
	})

	waitFor(t, 2*time.Second, func() bool {
		_, ok, _ := manager.GetState(context.Background(), clientID, "i-unsupported")
		return !ok
	})
}
