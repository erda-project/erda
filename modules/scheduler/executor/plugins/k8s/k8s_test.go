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

package k8s

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"gotest.tools/assert"
	apiv1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/engines"
	"github.com/erda-project/erda/pkg/istioctl/executors"
)

func TestComposeDeploymentNodeAffinityPreferredWithServiceWorkspace(t *testing.T) {
	k := Kubernetes{}
	workspace := "DEV"

	deploymentPreferred := []apiv1.PreferredSchedulingTerm{
		{
			Weight: 60,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-test",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 80,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-staging",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 100,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-prod",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 100,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/stateful-service",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
	}

	resPreferred := k.composeDeploymentNodeAntiAffinityPreferred(workspace)
	for index, preferred := range deploymentPreferred {
		assert.DeepEqual(t, preferred.Preference.MatchExpressions[0].Key, resPreferred[index].Preference.MatchExpressions[0].Key)
		assert.DeepEqual(t, preferred.Preference.MatchExpressions[0].Operator, resPreferred[index].Preference.MatchExpressions[0].Operator)
		assert.DeepEqual(t, preferred.Weight, resPreferred[index].Weight)
	}

}

func TestComposeStatefulSetNodeAffinityPreferredWithServiceWorkspace(t *testing.T) {
	k := Kubernetes{}
	workspace := "PROD"

	statefulSetPreferred := []apiv1.PreferredSchedulingTerm{
		{
			Weight: 60,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-dev",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 60,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-test",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 80,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/workspace-staging",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
		{
			Weight: 100,
			Preference: apiv1.NodeSelectorTerm{
				MatchExpressions: []apiv1.NodeSelectorRequirement{
					{
						Key:      "dice/stateless-service",
						Operator: apiv1.NodeSelectorOpDoesNotExist,
					},
				},
			},
		},
	}
	resPreferred := k.composeStatefulSetNodeAntiAffinityPreferred(workspace)
	for index, preferred := range statefulSetPreferred {
		assert.DeepEqual(t, preferred.Preference.MatchExpressions[0].Key, resPreferred[index].Preference.MatchExpressions[0].Key)
		assert.DeepEqual(t, preferred.Preference.MatchExpressions[0].Operator, resPreferred[index].Preference.MatchExpressions[0].Operator)
		assert.DeepEqual(t, preferred.Weight, resPreferred[index].Weight)
	}
}

func Test_getIstioEngine(t *testing.T) {
	mockEngine := &engines.LocalEngine{
		DefaultEngine: istioctl.NewDefaultEngine(&executors.AuthNExecutor{}),
	}
	type args struct {
		clusterName string
		info        apistructs.ClusterInfoData
	}
	tests := []struct {
		name    string
		args    args
		want    istioctl.IstioEngine
		wantErr bool
	}{
		{
			"case1",
			args{
				clusterName: "exist",
				info: apistructs.ClusterInfoData{
					apistructs.ISTIO_INSTALLED: "true",
				},
			},
			mockEngine,
			false,
		},
		{
			"case2",
			args{
				clusterName: "notExist",
				info: apistructs.ClusterInfoData{
					apistructs.ISTIO_INSTALLED: "true",
				},
			},
			istioctl.EmptyEngine,
			true,
		},
	}
	patch := monkey.Patch(engines.NewLocalEngine, func(clusterName string) (*engines.LocalEngine, error) {
		if clusterName == "exist" {
			return mockEngine, nil
		}
		return nil, errors.New("")
	})
	defer patch.Unpatch()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIstioEngine(tt.args.clusterName, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIstioEngine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getIstioEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}
