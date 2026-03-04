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
	"testing"
	"time"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func newTestManager(store state_store.LBStateStore, baseURL string) *Manager {
	return NewManager(store, Config{
		Probe: ProbeConfig{
			BaseURL:      baseURL,
			UnhealthyTTL: time.Hour,
			Timeout:      2 * time.Second,
		},
		Rescue: RescueConfig{
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     20 * time.Millisecond,
		},
	})
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
