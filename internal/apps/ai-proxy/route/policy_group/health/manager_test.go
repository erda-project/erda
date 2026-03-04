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
	"testing"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
)

func TestFilterHealthyInstances(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	manager := newTestManager(store, "http://127.0.0.1:65530")
	clientID := "client-a"

	unhealthy := ModelHealthState{
		State:     stateUnhealthy,
		APIType:   APITypeChatCompletions,
		UpdatedAt: time.Now(),
	}
	data, err := json.Marshal(unhealthy)
	if err != nil {
		t.Fatalf("marshal unhealthy state failed: %v", err)
	}
	if err := store.SetBinding(context.Background(), modelHealthBindingKey, makeModelHealthBindingID(clientID, "i1"), string(data), time.Hour); err != nil {
		t.Fatalf("set unhealthy state failed: %v", err)
	}

	instances := []*policygroup.RoutingModelInstance{
		testRoutingInstance("i1"),
		testRoutingInstance("i2"),
	}
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	filtered := manager.FilterHealthyInstances(policygroup.RouteRequest{Ctx: ctx, ClientID: clientID}, instances)
	if len(filtered) != 1 || filtered[0].ModelWithProvider.Id != "i2" {
		t.Fatalf("expected only i2 after health filter, got %v", collectIDs(filtered))
	}
	metaVal, ok := ctxhelper.GetPolicyGroupHealthMeta(ctx)
	if !ok || metaVal == nil {
		t.Fatal("expected policy group health meta in ctx")
	}
	meta, ok := metaVal.(*PolicyGroupHealthMeta)
	if !ok || meta == nil {
		t.Fatal("expected policy group health meta in ctx")
	}
	if len(meta.FilteredUnhealthyInstanceIDs) != 1 || meta.FilteredUnhealthyInstanceIDs[0] != "i1" {
		t.Fatalf("expected filtered unhealthy instance id i1, got %#v", meta.FilteredUnhealthyInstanceIDs)
	}

	probeCtx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutTrustedHealthProbe(probeCtx, true)
	probeReq := policygroup.RouteRequest{Ctx: probeCtx, ClientID: clientID}
	filteredProbe := manager.FilterHealthyInstances(probeReq, instances)
	if len(filteredProbe) != 2 {
		t.Fatalf("expected probe request to bypass health filter, got %v", collectIDs(filteredProbe))
	}
}

func TestFilterUnhealthyUnsupportedAPITypeReleasedAndRecorded(t *testing.T) {
	t.Parallel()

	store := state_store.NewMemoryStateStore()
	manager := newTestManager(store, "http://127.0.0.1:1")
	clientID := "client-a"

	stateBytes, err := json.Marshal(&ModelHealthState{
		State:     stateUnhealthy,
		APIType:   APIType("embeddings"),
		LastError: "read tcp: connection reset by peer",
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("marshal unhealthy state failed: %v", err)
	}
	if err := store.SetBinding(context.Background(), modelHealthBindingKey, makeModelHealthBindingID(clientID, "i1"), string(stateBytes), time.Hour); err != nil {
		t.Fatalf("set unhealthy state failed: %v", err)
	}

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	filtered := manager.FilterHealthyInstances(policygroup.RouteRequest{
		Ctx:      ctx,
		ClientID: clientID,
	}, []*policygroup.RoutingModelInstance{
		testRoutingInstance("i1"),
		testRoutingInstance("i2"),
	})
	if len(filtered) != 2 {
		t.Fatalf("expected unsupported api_type instance released, got %v", collectIDs(filtered))
	}
	if _, ok, _ := manager.GetState(context.Background(), clientID, "i1"); ok {
		t.Fatal("expected unhealthy state deleted for unsupported api_type")
	}
	metaVal, ok := ctxhelper.GetPolicyGroupHealthMeta(ctx)
	if !ok || metaVal == nil {
		t.Fatal("expected policy group health meta in ctx")
	}
	meta, ok := metaVal.(*PolicyGroupHealthMeta)
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

func TestMarkUnhealthyWritesModelMarkHeaderContext(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	manager := newTestManager(store, "http://127.0.0.1:1")

	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	manager.MarkUnhealthy(ctx, "client-a", "m-ctx", APITypeChatCompletions, "network timeout", nil)
	got, ok := ctxhelper.GetModelMarkUnhealthyInstanceID(ctx)
	if !ok || got != "m-ctx" {
		t.Fatalf("expected model mark unhealthy instance id m-ctx, got %q", got)
	}
}

func TestFilterHealthyInstancesIsolatedByClient(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	manager := newTestManager(store, "http://127.0.0.1:65530")

	unhealthy := ModelHealthState{
		State:     stateUnhealthy,
		APIType:   APITypeChatCompletions,
		UpdatedAt: time.Now(),
	}
	data, err := json.Marshal(unhealthy)
	if err != nil {
		t.Fatalf("marshal unhealthy state failed: %v", err)
	}
	if err := store.SetBinding(context.Background(), modelHealthBindingKey, makeModelHealthBindingID("client-a", "i1"), string(data), time.Hour); err != nil {
		t.Fatalf("set unhealthy state failed: %v", err)
	}

	instances := []*policygroup.RoutingModelInstance{
		testRoutingInstance("i1"),
		testRoutingInstance("i2"),
	}

	filteredA := manager.FilterHealthyInstances(policygroup.RouteRequest{
		Ctx:      ctxhelper.InitCtxMapIfNeed(context.Background()),
		ClientID: "client-a",
	}, instances)
	if len(filteredA) != 1 || filteredA[0].ModelWithProvider.Id != "i2" {
		t.Fatalf("expected client-a only i2 after health filter, got %v", collectIDs(filteredA))
	}

	filteredB := manager.FilterHealthyInstances(policygroup.RouteRequest{
		Ctx:      ctxhelper.InitCtxMapIfNeed(context.Background()),
		ClientID: "client-b",
	}, instances)
	if len(filteredB) != 2 {
		t.Fatalf("expected client-b not affected by client-a unhealthy mark, got %v", collectIDs(filteredB))
	}
}
