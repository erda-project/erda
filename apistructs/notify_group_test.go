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

package apistructs

import "testing"

func TestGetScopeDetail(t *testing.T) {
	testCases := []struct {
		name          string
		label         string
		wantScopeID   string
		wantScopeType string
	}{
		{
			name:          "empty label",
			label:         "",
			wantScopeID:   "",
			wantScopeType: "",
		},
		{
			name:          "project scope",
			label:         "{\"member_scopeID\":\"1\",\"member_scopeType\":\"project\"}",
			wantScopeID:   "1",
			wantScopeType: "project",
		},
		{
			name:          "wrong label",
			label:         "member_scopeID: 1",
			wantScopeType: "",
			wantScopeID:   "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n := &NotifyGroupDetail{
				Label: tc.label,
			}
			scopeID, scopeType := n.GetScopeDetail()
			if scopeID != tc.wantScopeID {
				t.Errorf("want scopeID: %s, but got: %s", tc.wantScopeID, scopeID)
			}
			if scopeType != tc.wantScopeType {
				t.Errorf("want scopeType: %s, but got: %s", tc.wantScopeType, scopeType)
			}
		})
	}
}
