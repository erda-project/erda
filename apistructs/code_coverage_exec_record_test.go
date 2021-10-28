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

func Test_checkWorkspace(t *testing.T) {
	type args struct {
		workspace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "check_workspace_success",
			args: args{
				workspace: DevEnv,
			},
			wantErr: false,
		},
		{
			name: "check_workspace_default",
			args: args{
				workspace: DefaultEnv,
			},
			wantErr: true,
		},
		{
			name: "check_workspace_fail",
			args: args{
				workspace: "",
			},
			wantErr: true,
		},
		{
			name: "check_workspace_fail",
			args: args{
				workspace: "asd",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkWorkspace(tt.args.workspace); (err != nil) != tt.wantErr {
				t.Errorf("checkWorkspace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
