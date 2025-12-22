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
	"context"
	"testing"

	pb "github.com/erda-project/erda-proto-go/apps/aiproxy/policy_group/pb"
)

func TestListAllOfficialLabels(t *testing.T) {
	h := &Handler{}
	resp, err := h.ListAllOfficialLabels(context.Background(), &pb.ListOfficialPolicyGroupLabelsRequest{})
	if err != nil {
		t.Fatalf("ListAllOfficialLabels error: %v", err)
	}
	if len(resp.LabelKeys) == 0 {
		t.Fatal("expected non-empty label keys")
	}
	// ensure sorted
	for i := 1; i < len(resp.LabelKeys); i++ {
		if resp.LabelKeys[i] < resp.LabelKeys[i-1] {
			t.Fatalf("label keys not sorted: %v", resp.LabelKeys)
		}
	}
}
