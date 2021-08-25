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

package endpoints

import "testing"

func Test_shouldCheckPermission(t *testing.T) {
	type args struct {
		isInternalClient       bool
		isInternalActionClient bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "isInternalClient",
			args: args{
				isInternalClient:       true,
				isInternalActionClient: false,
			},
			want: false,
		},
		{
			name: "isInternalActionClient",
			args: args{
				isInternalClient:       false,
				isInternalActionClient: true,
			},
			want: true,
		},
		{
			name: "isInternalClient_and_isInternalActionClient",
			args: args{
				isInternalClient:       true,
				isInternalActionClient: true,
			},
			want: true,
		},
		{
			name: "otherClient",
			args: args{
				isInternalClient:       false,
				isInternalActionClient: false,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCheckPermission(tt.args.isInternalClient, tt.args.isInternalActionClient); got != tt.want {
				t.Errorf("shouldCheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
