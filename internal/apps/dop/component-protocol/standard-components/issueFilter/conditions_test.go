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

package issueFilter

import "testing"

func TestIsEmpty(t *testing.T) {
	testCases := []struct {
		name       string
		conditions FrontendConditions
		want       bool
	}{
		{
			name: "empty",
			conditions: FrontendConditions{
				Severities: make([]string, 0),
			},
			want: true,
		},
		{
			name: "not empty",
			conditions: FrontendConditions{
				AssigneeIDs: []string{"1"},
				Severities:  make([]string, 0),
			},
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.conditions.IsEmpty() != tc.want {
				t.Errorf("got %v, want %v", tc.conditions.IsEmpty(), tc.want)
			}
		})
	}
}
