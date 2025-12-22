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

package selector

import (
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
)

func TestMatchSelector(t *testing.T) {
	labels := map[string]string{
		"country": "JP",
		"tier":    "primary",
	}
	sel := &pb.PolicySelector{
		Requirements: []*pb.PolicyRequirement{
			{Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(), Label: &pb.LabelRequirement{Key: "country", Operator: common_types.PolicySelectorLabelOpIn.String(), Values: []string{"JP", "US"}}},
			{Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(), Label: &pb.LabelRequirement{Key: "tier", Operator: common_types.PolicySelectorLabelOpExists.String()}},
		},
	}
	if !matchSelector(sel, labels) {
		t.Fatalf("expected selector match")
	}

	sel.Requirements[0].Label.Values = []string{"US"}
	if matchSelector(sel, labels) {
		t.Fatalf("expected selector not match")
	}

	sel.Requirements[0].Label.Values = []string{"US"}
	sel.Requirements[0].Label.Operator = common_types.PolicySelectorLabelOpNotIn.String()
	if !matchSelector(sel, labels) {
		t.Fatalf("expected selector match with not_in")
	}

	sel.Requirements[0].Label.Operator = common_types.PolicySelectorLabelOpDoesNotExist.String()
	sel.Requirements[0].Label.Key = "region"
	if !matchSelector(sel, labels) {
		t.Fatalf("expected does_not_exist match")
	}
}

func TestMatchRequirement(t *testing.T) {
	labels := map[string]string{"k": "v"}

	if !matchRequirement(nil, labels) {
		t.Fatalf("nil requirement should match")
	}

	req := &pb.PolicyRequirement{
		Type: common_types.PolicyBranchSelectorRequirementTypeLabel.String(),
		Label: &pb.LabelRequirement{
			Key:      "k",
			Operator: common_types.PolicySelectorLabelOpExists.String(),
		},
	}
	if !matchRequirement(req, labels) {
		t.Fatalf("label requirement should match")
	}

	if matchRequirement(&pb.PolicyRequirement{Type: "unknown"}, labels) {
		t.Fatalf("unknown type should not match")
	}
}

func TestMatchLabelRequirement(t *testing.T) {
	tests := []struct {
		name   string
		lr     *pb.LabelRequirement
		labels map[string]string
		want   bool
	}{
		{
			name: "nil requirement matches",
			lr:   nil,
			want: true,
		},
		{
			name: "in operator matches case insensitive key and value",
			lr: &pb.LabelRequirement{
				Key:      "Country",
				Operator: common_types.PolicySelectorLabelOpIn.String(),
				Values:   []string{"jp"},
			},
			labels: map[string]string{"country": "JP"},
			want:   true,
		},
		{
			name: "in operator missing label",
			lr: &pb.LabelRequirement{
				Key:      "country",
				Operator: common_types.PolicySelectorLabelOpIn.String(),
				Values:   []string{"JP"},
			},
			labels: map[string]string{"region": "APAC"},
			want:   false,
		},
		{
			name: "not in operator with present label",
			lr: &pb.LabelRequirement{
				Key:      "role",
				Operator: common_types.PolicySelectorLabelOpNotIn.String(),
				Values:   []string{"admin"},
			},
			labels: map[string]string{"role": "admin"},
			want:   false,
		},
		{
			name: "not in operator without label",
			lr: &pb.LabelRequirement{
				Key:      "role",
				Operator: common_types.PolicySelectorLabelOpNotIn.String(),
				Values:   []string{"admin"},
			},
			labels: map[string]string{"tier": "primary"},
			want:   true,
		},
		{
			name: "exists",
			lr: &pb.LabelRequirement{
				Key:      "tier",
				Operator: common_types.PolicySelectorLabelOpExists.String(),
			},
			labels: map[string]string{"tier": "primary"},
			want:   true,
		},
		{
			name: "does not exist",
			lr: &pb.LabelRequirement{
				Key:      "region",
				Operator: common_types.PolicySelectorLabelOpDoesNotExist.String(),
			},
			labels: map[string]string{"tier": "primary"},
			want:   true,
		},
		{
			name: "unknown operator",
			lr: &pb.LabelRequirement{
				Key:      "tier",
				Operator: "unknown",
			},
			labels: map[string]string{"tier": "primary"},
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := matchLabelRequirement(tt.lr, tt.labels); got != tt.want {
				t.Fatalf("matchLabelRequirement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrInSliceCaseInsensitive(t *testing.T) {
	if !strInSliceCaseInsensitive("JP", []string{"jp", "us"}) {
		t.Fatalf("expected to find value ignoring case")
	}
	if strInSliceCaseInsensitive("cn", []string{"jp", "us"}) {
		t.Fatalf("expected value not found")
	}
}

func TestLabelSelectorRequirement(t *testing.T) {
	selector := BuildLabelSelectorForKVIn("region", "us")
	if selector == nil || len(selector.Requirements) != 1 {
		t.Fatalf("expected single requirement")
	}
	req := selector.Requirements[0]
	if req.Type != common_types.PolicyBranchSelectorRequirementTypeLabel.String() {
		t.Fatalf("unexpected requirement type: %s", req.Type)
	}
	if req.Label == nil || req.Label.Key != "region" || req.Label.Operator != common_types.PolicySelectorLabelOpIn.String() {
		t.Fatalf("unexpected label requirement: %+v", req.Label)
	}
	if len(req.Label.Values) != 1 || req.Label.Values[0] != "us" {
		t.Fatalf("unexpected label values: %+v", req.Label.Values)
	}
}
