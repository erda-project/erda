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

package actionmgr

import (
	"testing"

	"github.com/stretchr/testify/assert"

	extensionpb "github.com/erda-project/erda-proto-go/core/extension/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/action/pb"
)

func Test_transfer2ExtensionReq(t *testing.T) {
	type arg struct {
		specYml string
	}
	testCases := []struct {
		name    string
		arg     arg
		wantErr bool
		want    *extensionpb.ExtensionVersionCreateRequest
	}{
		{
			name: "valid spec.yml",
			arg: arg{
				specYml: `
name: custom-script
version: "1.0"
type: action
displayName: custom-script
category: custom_task
desc: custom-script
public: true
labels:
  autotest: true
  configsheet: true
  project_level_app: true
  eci_disable: true

supportedVersions: # Deprecated. Please use supportedErdaVersions instead.
  - ">= 3.5"
supportedErdaVersions:
  - ">= 1.0"

params:
  - name: command
    desc: ${{ i18n.params.command.desc }}
locale:
  zh-CN:
    desc: 运行自定义命令
    displayName: 自定义任务
    params.command.desc: 运行的命令
  en-US:
    desc: Run custom commands
    displayName: Custom task
    params.command.desc: Command
`,
			},
			wantErr: false,
			want: &extensionpb.ExtensionVersionCreateRequest{
				Name:    "custom-script",
				Version: "1.0",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extReq, err := transfer2ExtensionReq(&pb.PipelineActionSaveRequest{
				Spec:   tc.arg.specYml,
				Dice:   " ",
				Readme: " ",
			})
			if (err != nil) != tc.wantErr {
				t.Errorf("tc: %s want err : %v, but got: %v", tc.name, tc.wantErr, err != nil)
			}
			assert.Equal(t, tc.want.Name, extReq.Name)
			assert.Equal(t, tc.want.Version, extReq.Version)
		})
	}
}
