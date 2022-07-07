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

package list

import "testing"

func Test_appAuthrized(t *testing.T) {
	type args struct {
		selectedOption string
		isMyapp        bool
		isPublic       bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{"my", false, false},
			want: true,
		},
		{
			args: args{"all", true, false},
			want: true,
		},
		{
			args: args{"all", false, true},
			want: true,
		},
		{
			args: args{"all", false, false},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appAuthrized(tt.args.selectedOption, tt.args.isMyapp, tt.args.isPublic); got != tt.want {
				t.Errorf("appAuthrized() = %v, want %v", got, tt.want)
			}
		})
	}
}
