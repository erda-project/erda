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

package engine

import (
	"context"
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

func TestSticky_BindingPinsInstance(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	group := &pb.PolicyGroup{
		Name:      "g1",
		Mode:      common_types.PolicyGroupModeWeighted.String(),
		StickyKey: "x-request-id",
		Branches: []*pb.PolicyBranch{
			{Name: "b1", Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String()},
		},
	}

	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group:    group,
		Instances: []*policygroup.RoutingModelInstance{
			testInstance("i1", nil),
			testInstance("i2", nil),
		},
		Meta: policygroup.RequestMeta{Keys: map[string]string{"x-request-id": "req-1"}},
	}

	inst1, trace1, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route1 failed: %v", err)
	}
	inst2, trace2, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route2 failed: %v", err)
	}
	if inst1.ModelWithProvider.Id != inst2.ModelWithProvider.Id {
		t.Fatalf("expected sticky routing, got %s then %s", inst1.ModelWithProvider.Id, inst2.ModelWithProvider.Id)
	}
	// First route has sticky key but no existing binding yet, so it will fall back to normal routing.
	if trace1.Sticky == nil || !trace1.Sticky.FallbackFromSticky {
		t.Fatalf("expected first route fallbackFromSticky=true, got false")
	}
	// Second route should hit binding and return without fallback.
	if trace2.Sticky != nil && trace2.Sticky.FallbackFromSticky {
		t.Fatalf("expected second route fallbackFromSticky=false, got true")
	}
}

func TestSticky_BranchReuse(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	group := &pb.PolicyGroup{
		Name:      "g1",
		StickyKey: "x-request-id",
		Mode:      common_types.PolicyGroupModeWeighted.String(),
		Branches: []*pb.PolicyBranch{
			{
				Name:     "b1",
				Weight:   1,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: selector.BuildLabelSelectorForKVIn("branch", "b1"),
			},
			{
				Name:     "b2",
				Weight:   1,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: selector.BuildLabelSelectorForKVIn("branch", "b2"),
			},
		},
	}

	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group:    group,
		Instances: []*policygroup.RoutingModelInstance{
			testInstance("i1", map[string]string{"branch": "b1"}),
			testInstance("i2", map[string]string{"branch": "b2"}),
		},
		Meta: policygroup.RequestMeta{Keys: map[string]string{"x-request-id": "req-1"}},
	}

	_, trace1, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route1 failed: %v", err)
	}
	_, trace2, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route2 failed: %v", err)
	}
	if trace1.Branch.Name != trace2.Branch.Name {
		t.Fatalf("expected sticky branch reuse, got %s then %s", trace1.Branch.Name, trace2.Branch.Name)
	}
}

func TestSticky_BindingInvalidatedWhenInstanceGone(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	group := &pb.PolicyGroup{
		Name:      "g1",
		Mode:      common_types.PolicyGroupModeWeighted.String(),
		StickyKey: "x-request-id",
		Branches: []*pb.PolicyBranch{
			{
				Name:     "b1",
				Weight:   1,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: selector.BuildLabelSelectorForKVIn("branch", "b1"),
			},
		},
	}

	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group:    group,
		Instances: []*policygroup.RoutingModelInstance{
			testInstance("i1", map[string]string{"branch": "b1"}),
			testInstance("i2", map[string]string{"branch": "b1"}),
		},
		Meta: policygroup.RequestMeta{Keys: map[string]string{"x-request-id": "req-1"}},
	}

	inst1, _, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route1 failed: %v", err)
	}
	if inst1 == nil {
		t.Fatal("route1 returned nil instance")
	}

	// Remove the bound instance from available set to ensure sticky binding is invalidated.
	req.Instances = []*policygroup.RoutingModelInstance{
		testInstance("i1", map[string]string{"branch": "b1"}),
		testInstance("i2", map[string]string{"branch": "b1"}),
	}
	kept := make([]*policygroup.RoutingModelInstance, 0, 1)
	for _, inst := range req.Instances {
		if inst.ModelWithProvider.Id != inst1.ModelWithProvider.Id {
			kept = append(kept, inst)
		}
	}
	if len(kept) != 1 {
		t.Fatalf("expected 1 remaining instance after removing bound one, got %d", len(kept))
	}
	req.Instances = kept
	inst2, trace2, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route2 failed: %v", err)
	}
	if inst2.ModelWithProvider.Id == inst1.ModelWithProvider.Id {
		t.Fatalf("expected reroute to a different instance, got same %s", inst2.ModelWithProvider.Id)
	}
	if trace2.Sticky == nil || !trace2.Sticky.FallbackFromSticky {
		t.Fatalf("expected fallbackFromSticky=true when bound instance is gone")
	}
}

func TestSticky_WithHeaderPrefix(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	engine := NewEngine(store)

	group := &pb.PolicyGroup{
		Name:      "g1",
		Mode:      common_types.PolicyGroupModeWeighted.String(),
		StickyKey: "req.header.x-request-id",
		Branches: []*pb.PolicyBranch{
			{
				Name:     "b1",
				Weight:   5,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: selector.BuildLabelSelectorForKVIn("branch", "b1"),
			},
			{
				Name:     "b2",
				Weight:   5,
				Strategy: common_types.PolicyGroupBranchStrategyRoundRobin.String(),
				Selector: selector.BuildLabelSelectorForKVIn("branch", "b2"),
			},
		},
	}

	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group:    group,
		Instances: []*policygroup.RoutingModelInstance{
			testInstance("i1", map[string]string{"branch": "b1"}),
			testInstance("i2", map[string]string{"branch": "b2"}),
		},
		Meta: policygroup.RequestMeta{Keys: map[string]string{"x-request-id": "sticky-1"}},
	}

	inst1, trace1, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route1 failed: %v", err)
	}
	inst2, trace2, err := engine.Route(context.Background(), req)
	if err != nil {
		t.Fatalf("route2 failed: %v", err)
	}
	if inst1.ModelWithProvider.Id != inst2.ModelWithProvider.Id {
		t.Fatalf("expected sticky to pin instance, got %s then %s", inst1.ModelWithProvider.Id, inst2.ModelWithProvider.Id)
	}
	if trace1.Branch.Name != trace2.Branch.Name {
		t.Fatalf("expected same branch, got %s then %s", trace1.Branch.Name, trace2.Branch.Name)
	}
}
