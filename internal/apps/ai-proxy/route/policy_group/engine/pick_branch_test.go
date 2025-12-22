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
)

func TestPickBranchWeighted_PicksHeavierFirst(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group: &pb.PolicyGroup{
			Name: "g-weighted",
			Mode: common_types.PolicyGroupModeWeighted.String(),
		},
	}

	candidates := []BranchCandidate{
		{branch: &pb.PolicyBranch{Name: "b1", Weight: 1}, instances: []*policygroup.RoutingModelInstance{testInstance("i1", nil)}},
		{branch: &pb.PolicyBranch{Name: "b2", Weight: 3}, instances: []*policygroup.RoutingModelInstance{testInstance("i2", nil)}},
	}

	picked, err := pickBranchWeighted(context.Background(), store, req, candidates)
	if err != nil {
		t.Fatalf("pickBranchWeighted failed: %v", err)
	}
	if picked == nil || picked.branch == nil {
		t.Fatalf("expected picked branch, got nil")
	}
	if picked.branch.Name != "b2" {
		t.Fatalf("expected b2 first, got %s", picked.branch.Name)
	}
}

func TestPickBranchByPriority_PicksLowestPriority(t *testing.T) {
	store := state_store.NewMemoryStateStore()
	req := policygroup.RouteRequest{
		ClientID: "c1",
		Group: &pb.PolicyGroup{
			Name: "g-priority",
			Mode: common_types.PolicyGroupModePriority.String(),
		},
	}

	candidates := []BranchCandidate{
		{branch: &pb.PolicyBranch{Name: "p1", Priority: 1}, instances: []*policygroup.RoutingModelInstance{testInstance("i1", nil)}},
		{branch: &pb.PolicyBranch{Name: "p2", Priority: 2}, instances: []*policygroup.RoutingModelInstance{testInstance("i2", nil)}},
	}

	picked, err := pickBranchByPriority(context.Background(), store, req, candidates)
	if err != nil {
		t.Fatalf("pickBranchByPriority failed: %v", err)
	}
	if picked == nil || picked.branch == nil {
		t.Fatalf("expected picked branch, got nil")
	}
	if picked.branch.Name != "p1" {
		t.Fatalf("expected lowest priority branch p1, got %s", picked.branch.Name)
	}
}
