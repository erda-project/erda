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

package handler_policy_group

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
)

func Test_buildBranchesFromInstances(t *testing.T) {
	branches := buildBranchesFromInstances([]*policypb.PolicyGroupInstanceWeight{
		{ModelInstanceId: "ins-1", Weight: 3},
		{ModelInstanceId: "ins-2", Weight: 1},
	})

	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
	first := branches[0]
	if first.Name != "ins-1" || first.Weight != 3 {
		t.Fatalf("unexpected first branch name/weight: %s/%d", first.Name, first.Weight)
	}
	if first.Priority != 1 || first.Strategy != common_types.PolicyGroupBranchStrategyRoundRobin.String() {
		t.Fatalf("unexpected branch priority/strategy: %d/%s", first.Priority, first.Strategy)
	}
	reqs := first.GetSelector().GetRequirements()
	if len(reqs) != 1 {
		t.Fatalf("expected 1 selector requirement, got %d", len(reqs))
	}
	req := reqs[0]
	if req.Type != common_types.PolicyBranchSelectorRequirementTypeLabel.String() {
		t.Fatalf("unexpected requirement type: %s", req.Type)
	}
	if req.Label == nil || req.Label.Key != common_types.PolicyLabelKeyModelInstanceID {
		t.Fatalf("unexpected label selector: %+v", req.Label)
	}
	if len(req.Label.Values) != 1 || req.Label.Values[0] != "ins-1" {
		t.Fatalf("unexpected label values: %v", req.Label.Values)
	}
}

func Test_buildForModelTemplateResponse(t *testing.T) {
	now := timestamppb.New(time.Unix(1700000000, 0))
	group := &policypb.PolicyGroup{
		Id:        "pg-1",
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "tpl-1",
		Mode:      common_types.PolicyGroupModeWeighted.String(),
		StickyKey: common_types.StickyKeyPrefixFromReqHeader + "x-user",
		Branches: []*policypb.PolicyBranch{
			{Name: "ins-1", Weight: 2},
			{Name: "ins-2", Weight: 5},
		},
	}

	resp := buildForModelTemplateResponse(group)

	if resp.PolicyGroupName != "tpl-1" || resp.TemplateId != "tpl-1" {
		t.Fatalf("unexpected names: policyGroup=%s template=%s", resp.PolicyGroupName, resp.TemplateId)
	}
	if resp.StickyHeaderKey != "x-user" {
		t.Fatalf("sticky header key should trim prefix, got %s", resp.StickyHeaderKey)
	}
	if len(resp.Instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(resp.Instances))
	}
	if resp.Instances[1].ModelInstanceId != "ins-2" || resp.Instances[1].Weight != 5 {
		t.Fatalf("unexpected second instance: %+v", resp.Instances[1])
	}
	if resp.CreatedAt.AsTime() != now.AsTime() || resp.UpdatedAt.AsTime() != now.AsTime() {
		t.Fatalf("unexpected timestamps: created=%v updated=%v", resp.CreatedAt, resp.UpdatedAt)
	}
}

func Test_findInstanceIDFromSelector(t *testing.T) {
	selector := &policypb.PolicySelector{
		Requirements: []*policypb.PolicyRequirement{
			{
				Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
				Label: &policypb.LabelRequirement{
					Key:      "Model-Instance-ID",
					Operator: common_types.PolicySelectorLabelOpIn.String(),
					Values:   []string{"ins-123"},
				},
			},
		},
	}
	if id, ok := findInstanceIDFromSelector(selector); !ok || id != "ins-123" {
		t.Fatalf("expected to find instance id ins-123, got %s", id)
	}

	selector.Requirements[0].Label.Operator = common_types.PolicySelectorLabelOpNotIn.String()
	if _, ok := findInstanceIDFromSelector(selector); ok {
		t.Fatal("expected not to match when operator is not 'in'")
	}
}
