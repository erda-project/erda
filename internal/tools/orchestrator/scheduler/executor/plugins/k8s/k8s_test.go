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
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/pkg/errors"
	"gotest.tools/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	papb "github.com/erda-project/erda-proto-go/orchestrator/podscaler/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/events/eventtypes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/executortypes"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/clusterinfo"
	ds "github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/daemonset"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/deployment"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/event"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/job"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sservice"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/namespace"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/persistentvolume"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/persistentvolumeclaim"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/pod"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/resourceinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/secret"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/serviceaccount"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/statefulset"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/storageclass"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/instanceinfo"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/istioctl"
	"github.com/erda-project/erda/pkg/istioctl/engines"
	"github.com/erda-project/erda/pkg/istioctl/executors"
	"github.com/erda-project/erda/pkg/k8sclient"
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

	monkey.Patch(k8sclient.NewForRestConfig, func(*rest.Config, ...k8sclient.Option) (*k8sclient.K8sClient, error) {
		return &k8sclient.K8sClient{}, nil
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
	assert.NilError(t, err)
}

func TestKubernetes_DeployInEdgeCluster(t *testing.T) {
	type fields struct {
		name                       executortypes.Name
		clusterName                string
		cluster                    *apistructs.ClusterInfo
		options                    map[string]string
		addr                       string
		client                     *httpclient.HTTPClient
		k8sClient                  *k8sclient.K8sClient
		bdl                        *bundle.Bundle
		evCh                       chan *eventtypes.StatusEvent
		deploy                     *deployment.Deployment
		job                        *job.Job
		ds                         *ds.Daemonset
		namespace                  *namespace.Namespace
		service                    *k8sservice.Service
		pvc                        *persistentvolumeclaim.PersistentVolumeClaim
		pv                         *persistentvolume.PersistentVolume
		sts                        *statefulset.StatefulSet
		pod                        *pod.Pod
		secret                     *secret.Secret
		storageClass               *storageclass.StorageClass
		sa                         *serviceaccount.ServiceAccount
		ClusterInfo                *clusterinfo.ClusterInfo
		resourceInfo               *resourceinfo.ResourceInfo
		event                      *event.Event
		cpuSubscribeRatio          float64
		memSubscribeRatio          float64
		devCpuSubscribeRatio       float64
		devMemSubscribeRatio       float64
		testCpuSubscribeRatio      float64
		testMemSubscribeRatio      float64
		stagingCpuSubscribeRatio   float64
		stagingMemSubscribeRatio   float64
		cpuNumQuota                float64
		elasticsearchoperator      addon.AddonOperator
		redisoperator              addon.AddonOperator
		mysqloperator              addon.AddonOperator
		canaloperator              addon.AddonOperator
		daemonsetoperator          addon.AddonOperator
		sourcecovoperator          addon.AddonOperator
		instanceinfoSyncCancelFunc context.CancelFunc
		dbclient                   *instanceinfo.Client
		istioEngine                istioctl.IstioEngine
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Test_01",
			fields: fields{
				ClusterInfo: &clusterinfo.ClusterInfo{},
			},
			want: false,
		},
		{
			name: "Test_02",
			fields: fields{
				ClusterInfo: &clusterinfo.ClusterInfo{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{
				name:                       tt.fields.name,
				clusterName:                tt.fields.clusterName,
				cluster:                    tt.fields.cluster,
				options:                    tt.fields.options,
				addr:                       tt.fields.addr,
				client:                     tt.fields.client,
				k8sClient:                  tt.fields.k8sClient,
				bdl:                        tt.fields.bdl,
				evCh:                       tt.fields.evCh,
				deploy:                     tt.fields.deploy,
				job:                        tt.fields.job,
				ds:                         tt.fields.ds,
				namespace:                  tt.fields.namespace,
				service:                    tt.fields.service,
				pvc:                        tt.fields.pvc,
				pv:                         tt.fields.pv,
				sts:                        tt.fields.sts,
				pod:                        tt.fields.pod,
				secret:                     tt.fields.secret,
				storageClass:               tt.fields.storageClass,
				sa:                         tt.fields.sa,
				ClusterInfo:                tt.fields.ClusterInfo,
				resourceInfo:               tt.fields.resourceInfo,
				event:                      tt.fields.event,
				cpuSubscribeRatio:          tt.fields.cpuSubscribeRatio,
				memSubscribeRatio:          tt.fields.memSubscribeRatio,
				devCpuSubscribeRatio:       tt.fields.devCpuSubscribeRatio,
				devMemSubscribeRatio:       tt.fields.devMemSubscribeRatio,
				testCpuSubscribeRatio:      tt.fields.testCpuSubscribeRatio,
				testMemSubscribeRatio:      tt.fields.testMemSubscribeRatio,
				stagingCpuSubscribeRatio:   tt.fields.stagingCpuSubscribeRatio,
				stagingMemSubscribeRatio:   tt.fields.stagingMemSubscribeRatio,
				cpuNumQuota:                tt.fields.cpuNumQuota,
				elasticsearchoperator:      tt.fields.elasticsearchoperator,
				redisoperator:              tt.fields.redisoperator,
				mysqloperator:              tt.fields.mysqloperator,
				canaloperator:              tt.fields.canaloperator,
				daemonsetoperator:          tt.fields.daemonsetoperator,
				sourcecovoperator:          tt.fields.sourcecovoperator,
				instanceinfoSyncCancelFunc: tt.fields.instanceinfoSyncCancelFunc,
				dbclient:                   tt.fields.dbclient,
				istioEngine:                tt.fields.istioEngine,
			}

			patch := monkey.PatchInstanceMethod(reflect.TypeOf(k.ClusterInfo), "Get", func(_ *clusterinfo.ClusterInfo) (map[string]string, error) {
				ret := make(map[string]string)
				if tt.name == "Test_01" {
					ret[string(apistructs.DICE_IS_EDGE)] = "false"
					return ret, nil
				}
				ret[string(apistructs.DICE_IS_EDGE)] = "true"
				return ret, nil
			})
			defer patch.Unpatch()
			if got := k.DeployInEdgeCluster(); got != tt.want {
				t.Errorf("DeployInEdgeCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runtimeIDMatch(t *testing.T) {
	type args struct {
		podRuntimeID string
		pod          apiv1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_01",
			args: args{
				podRuntimeID: "1",
				pod: apiv1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"DICE_RUNTIME": "1"},
					},
					Spec:   apiv1.PodSpec{},
					Status: apiv1.PodStatus{},
				},
			},
			want: true,
		},
		{
			name: "Test_02",
			args: args{
				podRuntimeID: "1",
				pod: apiv1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"DICE_RUNTIME_ID": "1"},
					},
					Spec:   apiv1.PodSpec{},
					Status: apiv1.PodStatus{},
				},
			},
			want: true,
		},
		{
			name: "Test_03",
			args: args{
				podRuntimeID: "1",
				pod: apiv1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"DICE_RUNTIME_XX": "1"},
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Env: []apiv1.EnvVar{
									{
										Name:  "DICE_RUNTIME",
										Value: "1",
									},
								},
							},
						},
					},
					Status: apiv1.PodStatus{},
				},
			},
			want: true,
		},
		{
			name: "Test_04",
			args: args{
				podRuntimeID: "1",
				pod: apiv1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"DICE_RUNTIME_XX": "1"},
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Env: []apiv1.EnvVar{
									{
										Name:  "DICE_RUNTIME_ID",
										Value: "1",
									},
								},
							},
						},
					},
					Status: apiv1.PodStatus{},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runtimeIDMatch(tt.args.podRuntimeID, tt.args.pod); got != tt.want {
				t.Errorf("runtimeIDMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ConvertToKedaScaledObject(t *testing.T) {
	scaled := papb.ScaledConfig{
		OrgID:    1234,
		RuleName: "test-rule",
		RuleID:   "",
		ScaleTargetRef: &papb.ScaleTargetRef{
			Name:       "test-target",
			Kind:       "test-kind",
			ApiVersion: "v2",
		},
		Advanced: &papb.HPAAdvanced{
			RestoreToOriginalReplicaCount: false,
		},
		Fallback: &papb.FallBack{
			Replicas: 3,
		},
		Triggers: []*papb.ScaleTriggers{
			{
				Type:     "Test-Type",
				Metadata: make(map[string]string),
			},
		},
	}

	object := convertToKedaScaledObject(scaled)
	t.Logf("%v", object)
}
