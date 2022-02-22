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

package deployment_order

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

func TestInspectDeploymentStatusDetail(t *testing.T) {
	type args struct {
		DeploymentOrder *dbclient.DeploymentOrder
		StatusMap       apistructs.DeploymentOrderStatusMap
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "first-batch",
			args: args{
				DeploymentOrder: &dbclient.DeploymentOrder{},
				StatusMap: apistructs.DeploymentOrderStatusMap{
					"java-demo": apistructs.DeploymentOrderStatusItem{
						AppID:            1,
						DeploymentID:     1,
						DeploymentStatus: apistructs.DeploymentStatusInit,
						RuntimeID:        1,
					},
				},
			},
			want: "{\"java-demo\":{\"appId\":1,\"deploymentId\":1,\"deploymentStatus\":\"INIT\",\"runtimeId\":1}}",
		},
		{
			name: "status-appending",
			args: args{
				DeploymentOrder: &dbclient.DeploymentOrder{
					StatusDetail: "{\"go-demo\":{\"appId\":0,\"deploymentId\":0,\"deploymentStatus\":\"INIT\",\"runtimeId\":0}}",
				},
				StatusMap: apistructs.DeploymentOrderStatusMap{
					"java-demo": apistructs.DeploymentOrderStatusItem{
						AppID:            1,
						DeploymentID:     1,
						DeploymentStatus: apistructs.DeploymentStatusDeploying,
						RuntimeID:        1,
					},
				},
			},
			want: "{\"go-demo\":{\"appId\":0,\"deploymentId\":0,\"deploymentStatus\":\"INIT\",\"runtimeId\":0},\"" +
				"java-demo\":{\"appId\":1,\"deploymentId\":1,\"deploymentStatus\":\"DEPLOYING\",\"runtimeId\":1}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := inspectDeploymentStatusDetail(tt.args.DeploymentOrder, tt.args.StatusMap)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, tt.args.DeploymentOrder.StatusDetail)
		})
	}
}
