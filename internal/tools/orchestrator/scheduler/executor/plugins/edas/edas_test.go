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

package edas

import "testing"

func Test_Set_Annotations(t *testing.T) {
	type args struct {
		name    string
		envs    map[string]string
		svcSpec *ServiceSpec
		wantErr bool
	}

	svc := &ServiceSpec{
		Name:  "fake-app",
		Image: "busybox",
	}

	tests := []args{
		{
			name:    "spec nil",
			wantErr: true,
		},
		{
			name:    "annotations",
			svcSpec: svc,
			envs: map[string]string{
				"DICE_ORG_ID":     "org-id",
				"DICE_RUNTIME_ID": "runtime-id",
			},
		},
		{
			name:    "empty annotations",
			svcSpec: svc,
			envs: map[string]string{
				"FAKE_KEY": "fake-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setAnnotations(tt.svcSpec, tt.envs); (err != nil) != tt.wantErr {
				t.Errorf("SetAnnotations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
