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

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	policypb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group/selector"
)

func TestPreviewLabelGroups_GroupBy(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "a"}),
		newInstance("i2", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "a"}),
		newInstance("i3", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "b"}),
	}

	groups, err := previewLabelGroups(instances, nil, common_types.PolicyLabelKeyServiceProviderType, common_types.LabelPreviewOpGroupBy.String(), "")
	if err != nil {
		t.Fatalf("previewLabelGroups error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	// groups are sorted by label value
	if groups[0].DisplayName != common_types.PolicyLabelKeyServiceProviderType+" = a" || groups[0].InstanceCount != 2 {
		t.Fatalf("unexpected first group: %+v", groups[0])
	}
	assertSelector(t, groups[0].Selector, common_types.PolicyLabelKeyServiceProviderType, common_types.PolicySelectorLabelOpIn.String(), []string{"a"})

	if groups[1].DisplayName != common_types.PolicyLabelKeyServiceProviderType+" = b" || groups[1].InstanceCount != 1 {
		t.Fatalf("unexpected second group: %+v", groups[1])
	}
	assertSelector(t, groups[1].Selector, common_types.PolicyLabelKeyServiceProviderType, common_types.PolicySelectorLabelOpIn.String(), []string{"b"})
}

func TestPreviewLabelGroups_Filter(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "a"}),
		newInstance("i2", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "a"}),
		newInstance("i3", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "b"}),
	}

	groups, err := previewLabelGroups(instances, nil, common_types.PolicyLabelKeyServiceProviderType, common_types.LabelPreviewOpFilter.String(), "a")
	if err != nil {
		t.Fatalf("previewLabelGroups error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].InstanceCount != 2 || groups[0].DisplayName != common_types.PolicyLabelKeyServiceProviderType+" = a" {
		t.Fatalf("unexpected group data: %+v", groups[0])
	}
	assertSelector(t, groups[0].Selector, common_types.PolicyLabelKeyServiceProviderType, common_types.PolicySelectorLabelOpIn.String(), []string{"a"})
}

func TestPreviewLabelGroups_Split(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "a"}),
		newInstance("i2", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "b"}),
		newInstance("i3", map[string]string{}), // missing provider-type
	}

	groups, err := previewLabelGroups(instances, nil, common_types.PolicyLabelKeyServiceProviderType, common_types.LabelPreviewOpSplit.String(), "a")
	if err != nil {
		t.Fatalf("previewLabelGroups error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	hit, miss := groups[0], groups[1]
	if hit.InstanceCount != 1 || hit.DisplayName != common_types.PolicyLabelKeyServiceProviderType+" = a" {
		t.Fatalf("unexpected hit group: %+v", hit)
	}
	assertSelector(t, hit.Selector, common_types.PolicyLabelKeyServiceProviderType, common_types.PolicySelectorLabelOpIn.String(), []string{"a"})

	if miss.InstanceCount != 2 || miss.DisplayName != common_types.PolicyLabelKeyServiceProviderType+" != a" {
		t.Fatalf("unexpected miss group: %+v", miss)
	}
	assertSelector(t, miss.Selector, common_types.PolicyLabelKeyServiceProviderType, common_types.PolicySelectorLabelOpNotIn.String(), []string{"a"})
}

func TestPreviewLabelGroups_BaseSelector(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "a", "region": "jp"}),
		newInstance("i2", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "a", "region": "us"}),
		newInstance("i3", map[string]string{common_types.PolicyLabelKeyServiceProviderType: "b", "region": "jp"}),
	}
	baseSelector := selector.BuildLabelSelectorForKVIn("region", "jp")

	groups, err := previewLabelGroups(instances, baseSelector, common_types.PolicyLabelKeyServiceProviderType, common_types.LabelPreviewOpGroupBy.String(), "")
	if err != nil {
		t.Fatalf("previewLabelGroups error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups after base filter, got %d", len(groups))
	}
	if groups[0].DisplayName != common_types.PolicyLabelKeyServiceProviderType+" = a" || groups[0].InstanceCount != 1 {
		t.Fatalf("unexpected first group: %+v", groups[0])
	}
	if groups[1].DisplayName != common_types.PolicyLabelKeyServiceProviderType+" = b" || groups[1].InstanceCount != 1 {
		t.Fatalf("unexpected second group: %+v", groups[1])
	}
}

func TestPreviewGroupBy_EmptyLabelValue(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{"k": "A"}),
		newInstance("i2", map[string]string{"k": ""}),
		newInstance("i3", map[string]string{"other": "v"}), // missing target label
	}

	groups := previewGroupBy("k", instances)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	empty := groups[0]
	if empty.DisplayName != "k = <empty>" || empty.InstanceCount != 2 {
		t.Fatalf("unexpected empty group: %+v", empty)
	}
	assertSelector(t, empty.Selector, "k", common_types.PolicySelectorLabelOpDoesNotExist.String(), nil)

	nonEmpty := groups[1]
	if nonEmpty.DisplayName != "k = A" || nonEmpty.InstanceCount != 1 {
		t.Fatalf("unexpected non-empty group: %+v", nonEmpty)
	}
	assertSelector(t, nonEmpty.Selector, "k", common_types.PolicySelectorLabelOpIn.String(), []string{"A"})
}

func TestPreviewFilter_Match(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{"k": "value"}),
		newInstance("i2", map[string]string{"k": "other"}),
	}
	groups := previewFilter("k", "value", instances)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	g := groups[0]
	if g.DisplayName != "k = value" || g.InstanceCount != 1 {
		t.Fatalf("unexpected filter group: %+v", g)
	}
	assertSelector(t, g.Selector, "k", common_types.PolicySelectorLabelOpIn.String(), []string{"value"})
}

func TestPreviewFilter_NoMatch(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{"k": "v1"}),
	}
	if groups := previewFilter("k", "other", instances); groups != nil {
		t.Fatalf("expected nil groups when no match, got %+v", groups)
	}
}

func TestPreviewSplit_AllHitAndEmptyRest(t *testing.T) {
	instances := []*policygroup.RoutingModelInstance{
		newInstance("i1", map[string]string{"k": "target"}),
	}
	groups := previewSplit("k", "target", instances)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	hit := groups[0]
	miss := groups[1]
	if hit.InstanceCount != 1 || hit.DisplayName != "k = target" {
		t.Fatalf("unexpected hit group: %+v", hit)
	}
	assertSelector(t, hit.Selector, "k", common_types.PolicySelectorLabelOpIn.String(), []string{"target"})

	if miss.InstanceCount != 0 || miss.DisplayName != "k != target" {
		t.Fatalf("unexpected miss group: %+v", miss)
	}
	assertSelector(t, miss.Selector, "k", common_types.PolicySelectorLabelOpNotIn.String(), []string{"target"})
}

func newInstance(id string, labels map[string]string) *policygroup.RoutingModelInstance {
	return &policygroup.RoutingModelInstance{
		ModelWithProvider: &cachehelpers.ModelWithProvider{
			Model: &modelpb.Model{Id: id},
		},
		Labels: labels,
	}
}

func assertSelector(t *testing.T, sel *policypb.PolicySelector, wantKey, wantOp string, wantValues []string) {
	t.Helper()
	if sel == nil || len(sel.Requirements) != 1 {
		t.Fatalf("unexpected selector: %+v", sel)
	}
	req := sel.Requirements[0]
	if req.Type != common_types.PolicyBranchSelectorRequirementTypeLabel.String() {
		t.Fatalf("unexpected requirement type: %s", req.Type)
	}
	if req.Label == nil || req.Label.Key != wantKey || req.Label.Operator != wantOp {
		t.Fatalf("unexpected label selector: %+v", req.Label)
	}
	if len(wantValues) == 0 && len(req.Label.Values) > 0 {
		t.Fatalf("unexpected values: %v", req.Label.Values)
	}
	if len(wantValues) > 0 {
		if len(req.Label.Values) != len(wantValues) {
			t.Fatalf("unexpected values size, got %v want %v", req.Label.Values, wantValues)
		}
		for i, v := range wantValues {
			if req.Label.Values[i] != v {
				t.Fatalf("unexpected value at %d: got %s want %s", i, req.Label.Values[i], v)
			}
		}
	}
}

func TestBuildLabelSelector(t *testing.T) {
	cases := []struct {
		name       string
		key        string
		op         string
		value      string
		wantValues []string
	}{
		{
			name:       "in with value",
			key:        "k",
			op:         common_types.PolicySelectorLabelOpIn.String(),
			value:      "v",
			wantValues: []string{"v"},
		},
		{
			name:       "not in with value",
			key:        "k",
			op:         common_types.PolicySelectorLabelOpNotIn.String(),
			value:      "v",
			wantValues: []string{"v"},
		},
		{
			name:  "does not exist without value",
			key:   "k",
			op:    common_types.PolicySelectorLabelOpDoesNotExist.String(),
			value: "",
		},
		{
			name:  "in without value should not set values",
			key:   "k",
			op:    common_types.PolicySelectorLabelOpIn.String(),
			value: "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			sel := buildLabelSelector(tt.key, tt.op, tt.value)
			if sel == nil || len(sel.Requirements) != 1 {
				t.Fatalf("unexpected selector: %+v", sel)
			}
			req := sel.Requirements[0]
			if req.Type != common_types.PolicyBranchSelectorRequirementTypeLabel.String() {
				t.Fatalf("unexpected requirement type: %s", req.Type)
			}
			if req.Label == nil {
				t.Fatalf("label requirement missing")
			}
			if req.Label.Key != tt.key || req.Label.Operator != tt.op {
				t.Fatalf("unexpected label data: %+v", req.Label)
			}
			if len(req.Label.Values) != len(tt.wantValues) {
				t.Fatalf("unexpected values length, got %v want %v", req.Label.Values, tt.wantValues)
			}
			for i, v := range tt.wantValues {
				if req.Label.Values[i] != v {
					t.Fatalf("unexpected value at %d: got %s want %s", i, req.Label.Values[i], v)
				}
			}
		})
	}
}
