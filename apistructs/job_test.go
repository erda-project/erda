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

func TestIsHostMode(t *testing.T) {
	type arg struct {
		netWork PodNetwork
	}
	testCases := []struct {
		name string
		arg  arg
		want bool
	}{
		{
			name: "host mode",
			arg: arg{
				netWork: PodNetwork{
					"mode": "host",
				},
			},
			want: true,
		},
		{
			name: "empty network",
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.arg.netWork.IsHostMode()
			if got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}
