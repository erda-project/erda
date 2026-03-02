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
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestFilterHealthyInstances(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	manager := NewManager(store, Config{})

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
	filtered := manager.FilterHealthyInstances(policygroup.RouteRequest{}, instances)
	if len(filtered) != 1 || filtered[0].ModelWithProvider.Id != "i2" {
		t.Fatalf("expected only i2 after health filter, got %v", collectIDs(filtered))
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
		ProbeBaseURL:   server.URL,
		UnhealthyTTL:   time.Hour,
		HealthyTTL:     time.Minute,
		ProbeTimeout:   2 * time.Second,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     20 * time.Millisecond,
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
		state, ok, _ := manager.GetState(context.Background(), "i1")
		return ok && state != nil && state.State == stateHealthy
	})
}

func TestBuildProbeHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer t_x")
	headers.Set("AK-Token", "ak-token-x")
	headers.Set("X-Trace-Id", "trace-1")
	headers.Set(vars.XAIProxyHealthProbe, "true")
	probeHeaders := BuildProbeHeaders(headers)

	if probeHeaders.Get("Authorization") == "" || probeHeaders.Get("AK-Token") == "" {
		t.Fatalf("expected auth headers kept, got: %v", probeHeaders)
	}
	if probeHeaders.Get("X-Trace-Id") != "" {
		t.Fatalf("unexpected non-auth header copied: %v", probeHeaders)
	}
	if probeHeaders.Get(vars.XAIProxyHealthProbe) != "" {
		t.Fatalf("unexpected probe marker copied into probe headers: %v", probeHeaders)
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
