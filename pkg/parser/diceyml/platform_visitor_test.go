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

package diceyml

import (
	"reflect"
	"testing"
)

func TestVisitorPlatform(t *testing.T) {
	type args struct {
		diceYml      string
		platformInfo map[string]string
	}

	testCases := []struct {
		name    string
		args    args
		gotFunc func(object *Object) string
		want    string
		wantErr bool
	}{
		{
			name: "render endpoints domain",
			args: args{
				diceYml: `
services:
  app-demo:
    ports:
      - port: 5000
        expose: true
    resources:
      cpu: 0.5
      mem: 500
    endpoints:
    - domain: test-${platform.DICE_PROJECT_NAME}-${platform.DICE_CLUSTER_NAME}.*
    - domain: test-${platform.DICE_PROJECT_NAME}-${platform.DICE_PROJECT_NAME}-${platform.DICE_PROJECT_NAME}.*
    deployments:
      replicas: 1
`,
				platformInfo: map[string]string{
					"DICE_PROJECT_NAME": "demo",
					"DICE_CLUSTER_NAME": "local",
				},
			},
			gotFunc: func(object *Object) string {
				return object.Services["app-demo"].Endpoints[0].Domain
			},
			want:    "test-demo-local.*",
			wantErr: false,
		},
		{
			name: "render arch image",
			args: args{
				diceYml: `
services:
  app-demo:
    ports:
      - port: 5000
        expose: true
    resources:
      cpu: 0.5
      mem: 500
    image: registry/demo/${platform.DICE_ARCH}/busybox
    deployments:
      replicas: 1
`,
				platformInfo: map[string]string{
					"DICE_ARCH": "arm64",
				},
			},
			gotFunc: func(object *Object) string {
				return object.Services["app-demo"].Image
			},
			want:    "registry/demo/arm64/busybox",
			wantErr: false,
		},
		{
			name: "job render",
			args: args{
				diceYml: `
jobs:
  job-demo:
    image: registry/demo/${platform.DICE_ARCH}/busybox
    resources:
      cpu: 2
      mem: 4096
      disk: 1024
`,
				platformInfo: map[string]string{
					"DICE_ARCH": "amd64",
				},
			},
			gotFunc: func(object *Object) string {
				return object.Jobs["job-demo"].Image
			},
			want:    "registry/demo/amd64/busybox",
			wantErr: false,
		},
		{
			name: "failed render",
			args: args{
				diceYml: `
services:
  app-demo:
    ports:
      - port: 5000
        expose: true
    resources:
      cpu: 0.5
      mem: 500
    image: registry/demo/${platform.DICE_ARCH_DEMO}/busybox
    deployments:
      replicas: 1
`,
				platformInfo: map[string]string{
					"DICE_ARCH": "arm64",
				},
			},
			gotFunc: func(object *Object) string {
				return object.Services["app-demo"].Image
			},
			want:    "registry/demo/arm64/busybox",
			wantErr: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			dice, err := New([]byte(tt.args.diceYml), true,
				WithPlatformInfo(tt.args.platformInfo))
			if err != nil {
				if !tt.wantErr {
					t.Fatalf("create dice from dice.yml failed, err: %v", err)
				}
				t.Logf("New wantErr: %v", err)
				return
			}

			got := tt.gotFunc(dice.obj)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("PlatformVisitor got = %v, want %v", got, tt.want)
			}
		})
	}
}
