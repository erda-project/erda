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
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"gotest.tools/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/scheduler/executor/plugins/k8s/clusterinfo"
	"github.com/erda-project/erda/modules/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/engines"
	"github.com/erda-project/erda/pkg/istioctl/executors"
	k8sclientconfig "github.com/erda-project/erda/pkg/k8sclient/config"
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

func TestNew(t *testing.T) {
	var (
		bdl         = bundle.New()
		mockCluster = "mock-cluster"
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetCluster", func(_ *bundle.Bundle, _ string) (*apistructs.ClusterInfo, error) {
		return &apistructs.ClusterInfo{}, nil
	})

	monkey.Patch(k8sclientconfig.ParseManageConfig, func(_ string, _ *apistructs.ManageConfig) (*rest.Config, error) {
		return &rest.Config{}, nil
	})

	monkey.Patch(util.GetClient, func(_ string, _ *apistructs.ManageConfig) (string, *httpclient.HTTPClient, error) {
		return "localhost", httpclient.New(), nil
	})

	monkey.Patch(dbengine.Open, func(_ ...*dbengine.Conf) (*dbengine.DBEngine, error) {
		return &dbengine.DBEngine{}, nil
	})

	monkey.Patch(clusterinfo.New, func(_ string, _ ...clusterinfo.Option) (*clusterinfo.ClusterInfo, error) {
		return &clusterinfo.ClusterInfo{}, nil
	})

	defer monkey.UnpatchAll()

	_, err := New("MARATHONFORMOCKCLUSTER", mockCluster, map[string]string{})
	//assert.NilError(t, err)
	fmt.Println(err)
}

func TestSetFineGrainedCPU(t *testing.T) {
	tests := []struct {
		name           string
		requestCpu     float64
		maxCpu         float64
		ratio          int
		wantErr        bool
		wantRequestCPU string
		wantMaxCPU     string
	}{
		{
			name:    "test1_request_cpu_not_set",
			wantErr: true,
		},
		{
			name:       "test2_invalid_max_cpu",
			requestCpu: 0.5,
			maxCpu:     0.25,
			wantErr:    true,
		},
		{
			name:           "test3_ratio_with_max_cpu_not_set",
			ratio:          2,
			requestCpu:     0.5,
			wantErr:        false,
			wantMaxCPU:     "500m",
			wantRequestCPU: "250m",
		},
		{
			name:           "test3_ratio_with_max_cpu_set",
			ratio:          2,
			requestCpu:     0.5,
			maxCpu:         1,
			wantErr:        false,
			wantMaxCPU:     "1000m",
			wantRequestCPU: "500m",
		},
	}

	k := &Kubernetes{}
	k.SetCpuQuota(100)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subRatio := float64(tt.ratio)

			cpu := fmt.Sprintf("%.fm", tt.requestCpu*1000)
			maxCpu := fmt.Sprintf("%.fm", tt.maxCpu*1000)
			c := &apiv1.Container{
				Name: "test-container",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse(cpu),
					},
					Limits: apiv1.ResourceList{
						apiv1.ResourceCPU: resource.MustParse(maxCpu),
					},
				},
			}

			err := k.SetFineGrainedCPU(c, map[string]string{}, subRatio)
			if (err != nil) != tt.wantErr {
				t.Errorf("failed, error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.Equal(t, tt.wantRequestCPU, fmt.Sprintf("%vm", c.Resources.Requests.Cpu().MilliValue()))
				assert.Equal(t, tt.wantMaxCPU, fmt.Sprintf("%vm", c.Resources.Limits.Cpu().MilliValue()))
			}
		})
	}
}
