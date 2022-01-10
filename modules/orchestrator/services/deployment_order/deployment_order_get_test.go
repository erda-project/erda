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
)

func TestParseDeploymentOrderStatus(t *testing.T) {
	type args struct {
		DeploymentStatus apistructs.DeploymentOrderStatusMap
	}
	tests := []struct {
		name string
		args args
		want apistructs.DeploymentOrderStatus
	}{
		{
			name: "apps-1",
			args: args{
				DeploymentStatus: apistructs.DeploymentOrderStatusMap{
					"app-1": apistructs.DeploymentOrderStatusItem{
						DeploymentStatus: apistructs.DeploymentStatusWaiting,
					},
					"app-2": apistructs.DeploymentOrderStatusItem{
						DeploymentStatus: apistructs.DeploymentStatusOK,
					},
				},
			},
			want: apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusDeploying),
		},
		{
			name: "apps-2",
			args: args{
				DeploymentStatus: apistructs.DeploymentOrderStatusMap{
					"app-1": apistructs.DeploymentOrderStatusItem{
						DeploymentStatus: apistructs.DeploymentStatusFailed,
					},
					"app-2": apistructs.DeploymentOrderStatusItem{
						DeploymentStatus: apistructs.DeploymentStatusOK,
					},
				},
			},
			want: apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusFailed),
		},
		{
			name: "apps-3",
			args: args{
				DeploymentStatus: nil,
			},
			want: apistructs.DeploymentOrderStatus(orderStatusWaitDeploy),
		},
		{
			name: "apps-4",
			args: args{
				DeploymentStatus: apistructs.DeploymentOrderStatusMap{
					"app-1": apistructs.DeploymentOrderStatusItem{
						DeploymentStatus: apistructs.DeploymentStatusCanceling,
					},
					"app-2": apistructs.DeploymentOrderStatusItem{
						DeploymentStatus: apistructs.DeploymentStatusOK,
					},
				},
			},
			want: apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusCanceled),
		},
		{
			name: "apps-5",
			args: args{
				DeploymentStatus: apistructs.DeploymentOrderStatusMap{
					"app-1": apistructs.DeploymentOrderStatusItem{
						DeploymentStatus: apistructs.DeploymentStatusOK,
					},
				},
			},
			want: apistructs.DeploymentOrderStatus(apistructs.DeploymentStatusOK),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDeploymentOrderStatus(tt.args.DeploymentStatus)

			if tt.want != got {
				t.Errorf("parseDeploymentOrderStatus got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComposeApplicationsInfo(t *testing.T) {
	type args struct {
		Releases   []*apistructs.ReleaseGetResponseData
		Params     map[string]apistructs.DeploymentOrderParam
		AppsStatus apistructs.DeploymentOrderStatusMap
	}

	appStatus := apistructs.DeploymentOrderStatusMap{
		"app1": {
			DeploymentID:     10,
			DeploymentStatus: apistructs.DeploymentStatusDeploying,
		},
	}

	params := map[string]apistructs.DeploymentOrderParam{
		"app1": {
			Env: []apistructs.DeploymentOrderParamItem{
				{Key: "key1", Value: "value1"},
			},
			File: []apistructs.DeploymentOrderParamItem{
				{Key: "key2", Value: "value2"},
			},
		},
	}

	tests := []struct {
		name string
		args args
		want []*apistructs.ApplicationInfo
	}{
		{
			name: "pipeline",
			args: args{
				Releases: []*apistructs.ReleaseGetResponseData{
					{
						ReleaseID: "8d2385a088df415decdf6357147ed4a2",
						Diceyml:   "fake-diceyaml",
						Labels: map[string]string{
							"gitCommitId": "27504bb7cb788bee08a50612b97faea201c0efed",
							"gitBranch":   "master",
						},
						ApplicationName: "app1",
					},
				},
				Params:     params,
				AppsStatus: appStatus,
			},
			want: []*apistructs.ApplicationInfo{
				{
					Name:         "app1",
					DeploymentId: 10,
					Param: "[{\"key\":\"key1\",\"value\":\"value1\",\"configType\":\"ENV\"},{\"key\":\"key2\",\"" +
						"value\":\"value2\",\"configType\":\"FILE\"}]",
					Branch:   "master",
					CommitId: "27504bb7cb788bee08a50612b97faea201c0efed",
					DiceYaml: "fake-diceyaml",
					Status:   apistructs.DeploymentStatusDeploying,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := composeApplicationsInfo(tt.args.Releases, tt.args.Params, tt.args.AppsStatus)
			assert.NoError(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}
