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

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/queue"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/instanceinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/deployment"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/deployment_order"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/domain"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/instance"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/migration"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/resource"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/runtime"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/goroutinepool"
)

func TestGetContainers(t *testing.T) {
	e := &Endpoints{
		instanceinfoImpl: &instanceinfo.InstanceInfoImpl{},
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(e.instanceinfoImpl), "GetInstanceInfo", func(i *instanceinfo.InstanceInfoImpl, req apistructs.InstanceInfoRequest) (apistructs.InstanceInfoDataList, error) {
		res := apistructs.InstanceInfoDataList{}
		for i := 0; i < 5; i++ {
			infoData := apistructs.InstanceInfoData{
				Meta:        fmt.Sprintf("Meta - %v", i),
				TaskID:      fmt.Sprintf("TaskID - %v", i),
				ContainerID: fmt.Sprintf("ContainerID - %v", i),
				ContainerIP: "127.0.0.1",
				HostIP:      "127.0.0.1",
				Image:       fmt.Sprintf("Image - %v", i),
				CpuRequest:  float64(i),
				MemRequest:  i,
				Phase:       fmt.Sprintf("Phase - %v", i),
				ExitCode:    200,
				Message:     "success",
				StartedAt:   time.Now(),
				ServiceName: fmt.Sprintf("ServiceName - %v", i),
				Cluster:     fmt.Sprintf("Cluster - %v", i),
			}
			if i%2 == 0 {
				finish := time.Now()
				infoData.FinishedAt = &finish
			}
			res = append(res, infoData)
		}
		return res, nil
	})
	_, err := e.getContainers(apistructs.InstanceInfoRequest{})
	if err != nil {
		t.Errorf("getContainers error %v", err.Error())
	}
}

func TestParseMeta(t *testing.T) {
	var (
		metaNamespace     = "namespace"
		metaPodUid        = "5e352011-f819-4dbb-bfea-3060cb866b53"
		metaPodName       = "test-pod"
		metaContainerName = "test-container"
	)
	tests := []struct {
		name  string
		input string
		want  apistructs.K8sInstanceMetaInfo
	}{
		{
			name:  "empty",
			input: "",
			want:  apistructs.K8sInstanceMetaInfo{},
		},
		{
			name:  "no meta",
			input: "hello world",
			want:  apistructs.K8sInstanceMetaInfo{},
		},
		{
			name: "one meta",
			input: strings.Join([]string{
				fmt.Sprintf("%s=%s", apistructs.K8sNamespace, metaNamespace),
				fmt.Sprintf("%s=%s", apistructs.K8sPodUid, metaPodUid),
				fmt.Sprintf("%s=%s", apistructs.K8sPodName, metaPodName),
				fmt.Sprintf("%s=%s", apistructs.K8sContainerName, metaContainerName),
			}, ","),
			want: apistructs.K8sInstanceMetaInfo{
				PodUid:        metaPodUid,
				PodName:       metaPodName,
				PodNamespace:  metaNamespace,
				ContainerName: metaContainerName,
			},
		},
		{
			name:  "invalid meta",
			input: "hello world:meta1=value1:meta2=value2:",
			want:  apistructs.K8sInstanceMetaInfo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInstanceMeta(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInstanceMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndpoints_getPodStatusFromK8s(t *testing.T) {
	type fields struct {
		db               *dbclient.DBClient
		queue            *queue.PusherQueue
		bdl              *bundle.Bundle
		pool             *goroutinepool.GoroutinePool
		evMgr            *events.EventManager
		runtime          *runtime.Runtime
		deployment       *deployment.Deployment
		deploymentOrder  *deployment_order.DeploymentOrder
		domain           *domain.Domain
		addon            *addon.Addon
		resource         *resource.Resource
		encrypt          *encryption.EnvEncrypt
		instance         *instance.Instance
		migration        *migration.Migration
		releaseSvc       pb.ReleaseServiceServer
		scheduler        *scheduler.Scheduler
		instanceinfoImpl *instanceinfo.InstanceInfoImpl
	}
	type args struct {
		runtimeID   string
		serviceName string
	}

	containers := make([]apistructs.PodContainer, 0)
	containers = append(containers, apistructs.PodContainer{
		ContainerID:   "ef2e8841543c3d6922991d6755de5f626a0d8e6639c24d52dd46971fd7fc80ef",
		ContainerName: "test",
		Image:         "nginx:v1.20",
		Resource: apistructs.ContainerResource{
			MemRequest: 64,
			MemLimit:   128,
			CpuRequest: 0.01,
			CpuLimit:   0.02,
		},
		Message: "Ok",
	})

	pods := make([]apistructs.Pod, 0)
	pods = append(pods, apistructs.Pod{
		Uid:           "a645a65d-84fa-446f-b2c3-f062a0db9bc3",
		IPAddress:     "10.112.227.92",
		Host:          "10.0.6.51",
		Phase:         "Healthy",
		Message:       containers[0].Message,
		StartedAt:     "2022-08-02T11:40:15+08:00",
		UpdatedAt:     "2022-08-02T11:40:15+08:00",
		Service:       "test",
		ClusterName:   "local-cluster",
		PodName:       "test-b8e0431e45-6b745f9665-xn9c5",
		K8sNamespace:  "project-1-test",
		RestartCount:  0,
		PodContainers: containers,
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []apistructs.Pod
		wantErr bool
	}{
		{
			name: "Test_01",
			fields: fields{
				runtime: &runtime.Runtime{},
			},
			args: args{
				runtimeID:   "1",
				serviceName: "test",
			},
			want:    pods,
			wantErr: false,
		},
		{
			name: "Test_02",
			fields: fields{
				runtime: &runtime.Runtime{},
			},
			args: args{
				runtimeID:   "1",
				serviceName: "test2",
			},
			want:    []apistructs.Pod{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoints{
				db:               tt.fields.db,
				queue:            tt.fields.queue,
				bdl:              tt.fields.bdl,
				pool:             tt.fields.pool,
				evMgr:            tt.fields.evMgr,
				runtime:          tt.fields.runtime,
				deployment:       tt.fields.deployment,
				deploymentOrder:  tt.fields.deploymentOrder,
				domain:           tt.fields.domain,
				addon:            tt.fields.addon,
				resource:         tt.fields.resource,
				encrypt:          tt.fields.encrypt,
				instance:         tt.fields.instance,
				migration:        tt.fields.migration,
				releaseSvc:       tt.fields.releaseSvc,
				scheduler:        tt.fields.scheduler,
				instanceinfoImpl: tt.fields.instanceinfoImpl,
			}

			m1 := monkey.PatchInstanceMethod(reflect.TypeOf(e.runtime), "GetRuntimeServiceCurrentPods", func(rt *runtime.Runtime, runtimeID uint64, serviceName string) (*apistructs.ServiceGroup, error) {

				extra := make(map[string]string)
				extra["test"] = "[{\n    \"apiVersion\": \"v1\",\n    \"kind\": \"Pod\",\n    \"metadata\": {\n" +
					"\"creationTimestamp\": \"2022-08-02T03:40:15Z\",\n        \"labels\": {\n" +
					"\"DICE_CLUSTER_NAME\": \"local-cluster\",\n            \"app\": \"test\"\n        },\n" +
					"\"name\": \"test-b8e0431e45-6b745f9665-xn9c5\",\n        \"namespace\": \"project-1-test\",\n" +
					" \"resourceVersion\": \"13849435\",\n        \"uid\": \"a645a65d-84fa-446f-b2c3-f062a0db9bc3\"\n},\n" +
					"\"spec\": {\n        \"containers\": [\n            {\n                \"env\": [\n" +
					" {\n                        \"name\": \"DICE_CLUSTER_NAME\",\n \"value\": \"local-cluster\"\n}\n],\n" +
					" \"image\": \"nginx:v1.20\",\n                \"name\": \"test\",\n      \"resources\": {\n" +
					"                   \"limits\": {\n                        \"cpu\": \"20m\",\n" +
					"\"ephemeral-storage\": \"20Gi\",\n                        \"memory\": \"134217728\"\n" +
					"},\n                    \"requests\": {\n                        \"cpu\": \"10m\",\n" +
					"  \"ephemeral-storage\": \"1Gi\",\n                        \"memory\": \"67108864\"\n" +
					"}\n                },\n                \"terminationMessagePath\": \"/dev/termination-log\",\n" +
					" \"terminationMessagePolicy\": \"File\"\n            }\n        ],\n        \"dnsPolicy\": " +
					"\"ClusterFirst\",\n        \"enableServiceLinks\": false,\n        \"nodeName\": \"10.0.6.51\",\n" +
					" \"preemptionPolicy\": \"PreemptLowerPriority\",\n        \"priority\": 0,\n        \"restartPolicy\":" +
					" \"Always\",\n        \"schedulerName\": \"default-scheduler\",\n        \"securityContext\": {},\n" +
					"\"serviceAccount\": \"default\",\n        \"serviceAccountName\": \"default\",\n" +
					"\"shareProcessNamespace\": false,\n        \"terminationGracePeriodSeconds\": 30\n    },\n" +
					" \"status\": {\n        \"containerStatuses\": [\n            {\n                \"containerID\":" +
					" \"docker://ef2e8841543c3d6922991d6755de5f626a0d8e6639c24d52dd46971fd7fc80ef\",\n" +
					"\"image\": \"nginx:v1.20\",\n                \"name\": \"test\",\n" +
					"\"ready\": true,\n                \"restartCount\": 0,\n                \"started\": true,\n" +
					" \"state\": {\n                    \"running\": {\n" +
					"\"startedAt\": \"2022-08-02T03:40:16Z\"\n                    }\n                }\n" +
					"}\n        ],\n        \"hostIP\": \"10.0.6.51\",\n        \"phase\": \"Running\",\n" +
					"\"podIP\": \"10.112.227.92\",\n        \"podIPs\": [\n            {\n" +
					"\"ip\": \"10.112.227.92\"\n            }\n        ],\n        \"qosClass\": \"Burstable\",\n" +
					"\"startTime\": \"2022-08-02T03:40:15Z\"\n    }\n}]"

				ret := &apistructs.ServiceGroup{
					Extra: extra,
				}

				return ret, nil
			})

			defer m1.Unpatch()

			got, err := e.getPodStatusFromK8s(tt.args.runtimeID, tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPodStatusFromK8s() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPodStatusFromK8s() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertReasonMessageToHumanReadableMessage(t *testing.T) {
	type args struct {
		containerReason  string
		containerMessage string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "Test01",
			args: args{
				containerReason:  "Unschedulable",
				containerMessage: "0/6 nodes are available: 6 Insufficient memory.",
			},
			want: "Failed to Schedule Pod: Insufficient memory",
		},
		{
			name: "Test02",
			args: args{
				containerReason:  "Unschedulable",
				containerMessage: "0/6 nodes are available: 6 Insufficient cpu.",
			},
			want: "Failed to Schedule Pod: Insufficient cpu",
		},
		{
			name: "Test03",
			args: args{
				containerReason:  "Unschedulable",
				containerMessage: "0/6 nodes are available: 6 node(s) didn't match Pod's node affinity.",
			},
			want: "Failed to Schedule Pod: didn't match Pod's node affinity",
		},
		{
			name: "Test04",
			args: args{
				containerReason:  "RunContainerError",
				containerMessage: "failed to start container \"6be8ec962041b4f0d0afc515af35df4ea976130bb79d402f12514a7a2c2dd96f\": Error response from daemon: OCI runtime create failed: runc create failed: unable to start container process: exec: \"/opt/xxx xxx\": stat /opt/xxx xxx: no such file or directory: unknown",
			},
			want: "Failed to start container: some file or directory not found",
		},
		{
			name: "Test05",
			args: args{
				containerReason:  "CrashLoopBackOff",
				containerMessage: "back-off 5m0s restarting failed container=hpa pod=testvpa-5548879d76-79rtb_default(d0c3f1de-55b8-41ca-a402-3da9bff84f2a)",
			},
			want: "Failed to run command: container process exit",
		},
		{
			name: "Test06",
			args: args{
				containerReason:  "ImagePullBackOff",
				containerMessage: "rpc error: code = Unknown desc = Error response from daemon: manifest for registry.cn-hangzhou.aliyuncs.com/xxx/xa:v0.09 not found",
			},
			want: "Failed to pull image: image not found or can not pull",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertReasonMessageToHumanReadableMessage(tt.args.containerReason, tt.args.containerMessage); got != tt.want {
				t.Errorf("convertReasonMessageToHumanReadableMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMainContainerNameByServiceName(t *testing.T) {
	// Prepare test cases
	testCases := []struct {
		description    string
		containers     []corev1.Container
		serviceName    string
		expectedResult string
	}{
		{
			description: "Empty serviceName should return the name of the first container",
			containers: []corev1.Container{
				{Name: "container1"},
				{Name: "container2"},
			},
			serviceName:    "",
			expectedResult: "container1",
		},
		{
			description: "serviceName found in container's environment variables",
			containers: []corev1.Container{
				{Name: "container1", Env: []corev1.EnvVar{{Name: apistructs.EnvDiceServiceName, Value: "service1"}}},
				{Name: "container2", Env: []corev1.EnvVar{{Name: apistructs.EnvDiceServiceName, Value: "service2"}}},
			},
			serviceName:    "service2",
			expectedResult: "container2",
		},
		{
			description: "serviceName not found in container's environment variables, should return the name of the first container",
			containers: []corev1.Container{
				{Name: "container1", Env: []corev1.EnvVar{{Name: apistructs.EnvDiceServiceName, Value: "service1"}}},
				{Name: "container2", Env: []corev1.EnvVar{{Name: apistructs.EnvDiceServiceName, Value: "service2"}}},
			},
			serviceName:    "nonexistent-service",
			expectedResult: "container1",
		},
		{
			description: "Only one container, serviceName is empty, should return its name",
			containers: []corev1.Container{
				{Name: "single-container"},
			},
			serviceName:    "",
			expectedResult: "single-container",
		},
		{
			description: "Only one container, serviceName is not empty, should return its name",
			containers: []corev1.Container{
				{Name: "single-container"},
			},
			serviceName:    "my-service",
			expectedResult: "single-container",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := getMainContainerNameByServiceName(tc.containers, tc.serviceName)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestSortContainers(t *testing.T) {
	testCases := []struct {
		description     string
		containers      []apistructs.PodContainer
		mainServiceName string
		expectedResult  []apistructs.PodContainer
	}{
		{
			description:     "Empty containers list, should return empty list",
			containers:      []apistructs.PodContainer{},
			mainServiceName: "",
			expectedResult:  []apistructs.PodContainer{},
		},
		{
			description: "Only one container, should return the same list",
			containers: []apistructs.PodContainer{
				{ContainerName: "container1"},
			},
			mainServiceName: "container1",
			expectedResult: []apistructs.PodContainer{
				{ContainerName: "container1"},
			},
		},
		{
			description: "Main service found, should place it at the beginning of the list",
			containers: []apistructs.PodContainer{
				{ContainerName: "container1"},
				{ContainerName: "container2"},
				{ContainerName: "container3"},
			},
			mainServiceName: "container2",
			expectedResult: []apistructs.PodContainer{
				{ContainerName: "container2"},
				{ContainerName: "container1"},
				{ContainerName: "container3"},
			},
		},
		{
			description: "Main service not found, should return the same list",
			containers: []apistructs.PodContainer{
				{ContainerName: "container1"},
				{ContainerName: "container2"},
				{ContainerName: "container3"},
			},
			mainServiceName: "nonexistent-container",
			expectedResult: []apistructs.PodContainer{
				{ContainerName: "container1"},
				{ContainerName: "container2"},
				{ContainerName: "container3"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := sortContainers(tc.containers, tc.mainServiceName)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
