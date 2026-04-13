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

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	api "github.com/aliyun/alibaba-cloud-sdk-go/services/edas"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

const (
	fakeAppId         = "app-id-1"
	fakeAppName       = "app-name-1"
	fakeChangeOrderId = "change-order-1"
)

var fakeDeployment = &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "app-1",
		Namespace: metav1.NamespaceDefault,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: pointer.Int32(1),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "app-1",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "app-1",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "busybox",
					},
				},
			},
		},
	},
}

var fakeServiceSpec = &types.ServiceSpec{
	Name:        fakeAppName,
	Image:       "busybox",
	Cmd:         "sh",
	Args:        "-c \"sleep 1000\"",
	Instances:   1,
	CPU:         1,
	Mcpu:        1,
	Mem:         128,
	Ports:       []int{8080},
	Envs:        "[{\"name\":\"testkey\",\"value\":\"testValue\"}]",
	Liveness:    "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"tcpSocket\":{\"host\":\"\", \"port\":8080}}",
	Readiness:   "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"httpGet\": {\"path\": \"/consumer\",\"port\": 8080,\"scheme\": \"HTTP\",\"httpHeaders\": [{\"name\": \"test\",\"value\": \"testvalue\"}]}}",
	Annotations: "{\"annotation-name-1\":\"annotation-value-1\",\"annotation-name-2\":\"annotation-value-2\"}",
	Labels:      "{\"app\":\"test-service\",\"servicegroup-id\":\"test-sg-id\"}",
}

func mock(c *api.Client) *gomonkey.Patches {
	patches := gomonkey.NewPatches()

	mockDoAction := func(_ *sdk.Client, request requests.AcsRequest, response responses.AcsResponse) error {
		switch req := request.(type) {
		case *api.ListApplicationRequest:
			resp := response.(*api.ListApplicationResponse)
			resp.RequestId = uuid.New()
			if req.AppName == fakeAppName {
				resp.ApplicationList = api.ApplicationList{
					Application: []api.ApplicationInListApplication{
						{
							AppId:     fakeAppId,
							Name:      fakeAppName,
							ClusterId: "cluster-a",
						},
					},
				}
			}
			return nil
		case *api.StopK8sApplicationRequest:
			resp := response.(*api.StopK8sApplicationResponse)
			resp.RequestId = uuid.New()
			if req.AppId == fakeAppId {
				resp.ChangeOrderId = fakeChangeOrderId
			}
			return nil
		case *api.ListRecentChangeOrderRequest:
			resp := response.(*api.ListRecentChangeOrderResponse)
			resp.RequestId = uuid.New()
			if req.AppId == fakeAppId {
				resp.ChangeOrderList = api.ChangeOrderList{
					ChangeOrder: []api.ChangeOrder{
						{
							Status:        1,
							ChangeOrderId: fakeChangeOrderId,
							AppId:         fakeAppId,
						},
					},
				}
			}
			return nil
		case *api.AbortChangeOrderRequest:
			resp := response.(*api.AbortChangeOrderResponse)
			resp.RequestId = uuid.New()
			return nil
		case *api.DeleteK8sApplicationRequest:
			resp := response.(*api.DeleteK8sApplicationResponse)
			resp.RequestId = uuid.New()
			return nil
		case *api.GetChangeOrderInfoRequest:
			resp := response.(*api.GetChangeOrderInfoResponse)
			resp.RequestId = uuid.New()
			if req.ChangeOrderId == fakeChangeOrderId {
				resp.ChangeOrderInfo = api.ChangeOrderInfo{
					ChangeOrderId: fakeChangeOrderId,
					Status:        int(types.CHANGE_ORDER_STATUS_SUCC),
				}
			}
			return nil
		case *api.GetAppDeploymentRequest:
			resp := response.(*api.GetAppDeploymentResponse)
			resp.RequestId = uuid.New()
			if req.AppId == fakeAppId {
				deployJSON, err := json.Marshal(fakeDeployment)
				if err != nil {
					return err
				}
				resp.Data = string(deployJSON)
			}
			return nil
		case *api.ScaleK8sApplicationRequest:
			if req.AppId != fakeAppId {
				return errors.New("scale fail")
			}
			resp := response.(*api.ScaleK8sApplicationResponse)
			resp.RequestId = uuid.New()
			return nil
		case *api.DeployK8sApplicationRequest:
			if req.AppId != fakeAppId {
				return errors.New("deploy fail")
			}
			resp := response.(*api.DeployK8sApplicationResponse)
			resp.RequestId = uuid.New()
			resp.ChangeOrderId = fakeChangeOrderId
			return nil
		case *api.InsertK8sApplicationRequest:
			if req.AppName != fakeAppName {
				return errors.New("insert k8s application fail")
			}
			resp := response.(*api.InsertK8sApplicationResponse)
			resp.RequestId = uuid.New()
			resp.ApplicationInfo = api.ApplicationInfo{
				AppId:         fakeAppId,
				ChangeOrderId: fakeChangeOrderId,
			}
			return nil
		case *api.QueryApplicationStatusRequest:
			if req.AppId != fakeAppId {
				return errors.New("query application fail")
			}
			resp := response.(*api.QueryApplicationStatusResponse)
			resp.RequestId = uuid.New()
			resp.AppInfo = api.AppInfo{
				EccList: api.EccList{
					Ecc: []api.Ecc{
						{
							AppState:  int(types.APP_STATE_STOPPED),
							TaskState: int(types.TASK_STATE_SUCCESS),
						},
					},
				},
			}
			return nil
		default:
			return nil
		}
	}

	patches.ApplyMethod(reflect.TypeOf(&sdk.Client{}), "DoAction", mockDoAction)
	patches.ApplyMethod(reflect.TypeOf(c), "DoAction", func(_ *api.Client, request requests.AcsRequest, response responses.AcsResponse) error {
		return mockDoAction(nil, request, response)
	})

	return patches
}

func newTestWrapEDASClient() (Interface, *gomonkey.Patches) {
	client := &api.Client{}
	c := New(logrus.WithField("unit-test", "test"), client, "addr", "cluster-a",
		"cn-hangzhou", "cn-hangzhou:erda", "true")
	patches := mock(client)
	return c, patches
}

func TestGetAppID(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		appName string
	}

	tests := []struct {
		name     string
		args     args
		expected string
		wantErr  bool
	}{
		{
			name: "Test successful case with single application",
			args: args{
				appName: fakeAppName,
			},
			expected: fakeAppId,
			wantErr:  false,
		},
		{
			name: "Test case with no applications",
			args: args{
				appName: "no-existed-app-name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetAppID(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAppID error, got: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("Expected app ID: %s, got: %s", tt.expected, got)
			}
		})
	}
}

func TestDeleteAppByName(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		appName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test successful case with single application",
			args: args{
				appName: fakeAppName,
			},
			wantErr: false,
		},
		{
			name: "Test case with no applications",
			args: args{
				appName: "no-existed-app-name",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.DeleteAppByName(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteAppByName error, err: %v", err)
			}
		})
	}
}

func TestGetAppDeployment(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		appName string
	}

	tests := []struct {
		name     string
		args     args
		expected *appsv1.Deployment
		wantErr  bool
	}{
		{
			name: "Test successful case with single application",
			args: args{
				appName: fakeAppName,
			},
			expected: fakeDeployment,
			wantErr:  false,
		},
		{
			name: "Test case with no applications",
			args: args{
				appName: "no-existed-app-name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetAppDeployment(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAppDeployment error, got: %v", err)
			}

			if (got != nil && tt.expected != nil) && (got.Name != tt.expected.Name) {
				t.Fatalf("GetAppDeployment, expected: %v, got: %v", tt.expected.Name, got.Name)
			}
		})
	}
}

func TestScaleK8sApplication(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		appId    string
		replicas int
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test successful case with single application",
			args: args{
				appId:    fakeAppId,
				replicas: 2,
			},
			wantErr: false,
		},
		{
			name: "Test case with no applications",
			args: args{
				appId: "no-existed-app-name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.ScaleApp(tt.args.appId, tt.args.replicas)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ScaleApp error, got: %v", err)
			}
		})
	}
}

func TestDeployK8sApplication(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		appId string
		spec  *types.ServiceSpec
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test successful case with single application",
			args: args{
				appId: fakeAppId,
				spec: &types.ServiceSpec{
					Name:        "service-a",
					Image:       "busybox",
					Cmd:         "sh",
					Args:        "-c \"sleep 1000\"",
					Instances:   1,
					CPU:         1,
					Mcpu:        1,
					Mem:         128,
					Ports:       []int{8080},
					Envs:        "[{\"name\":\"testkey\",\"value\":\"testValue\"}]",
					Liveness:    "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"tcpSocket\":{\"host\":\"\", \"port\":8080}}",
					Readiness:   "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"httpGet\": {\"path\": \"/consumer\",\"port\": 8080,\"scheme\": \"HTTP\",\"httpHeaders\": [{\"name\": \"test\",\"value\": \"testvalue\"}]}}",
					Annotations: "{\"annotation-name-1\":\"annotation-value-1\",\"annotation-name-2\":\"annotation-value-2\"}",
					Labels:      "{\"app\":\"test-service\",\"servicegroup-id\":\"test-sg-id\",\"environment\":\"test\"}",
				},
			},
			wantErr: false,
		},
		{
			name: "Test case with no applications",
			args: args{
				appId: "no-existed-app-name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.DeployApp(tt.args.appId, tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeployApp error, got: %v", err)
			}
		})
	}
}

func TestInsertK8sApplicationResponse(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		spec *types.ServiceSpec
	}

	tests := []struct {
		name     string
		args     args
		excepted string
		wantErr  bool
	}{
		{
			name: "Test successful case with single application",
			args: args{
				spec: fakeServiceSpec,
			},
			excepted: fakeAppId,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.InsertK8sApp(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InsertK8sApp error, got: %v", err)
			}
			if got != tt.excepted {
				t.Fatalf("InsertK8sApp error, got: %v, want: %v", got, tt.excepted)
			}
		})
	}
}

func TestQueryAppStatus(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		appName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test successful case with single application",
			args: args{
				appName: fakeAppName,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.QueryAppStatus(tt.args.appName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("QueryAppStatus error, got: %v", err)
			}
		})
	}
}

func TestInsertK8sApplicationWithLabels(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		spec *types.ServiceSpec
	}

	tests := []struct {
		name     string
		args     args
		expected string
		wantErr  bool
	}{
		{
			name: "Test successful case with labels",
			args: args{
				spec: fakeServiceSpec,
			},
			expected: fakeAppId,
			wantErr:  false,
		},
		{
			name: "Test case with empty labels",
			args: args{
				spec: &types.ServiceSpec{
					Name:        fakeAppName,
					Image:       "busybox",
					Cmd:         "sh",
					Args:        "-c \"sleep 1000\"",
					Instances:   1,
					CPU:         1,
					Mcpu:        1,
					Mem:         128,
					Ports:       []int{8080},
					Envs:        "[{\"name\":\"testkey\",\"value\":\"testValue\"}]",
					Liveness:    "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"tcpSocket\":{\"host\":\"\", \"port\":8080}}",
					Readiness:   "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"httpGet\": {\"path\": \"/consumer\",\"port\": 8080,\"scheme\": \"HTTP\",\"httpHeaders\": [{\"name\": \"test\",\"value\": \"testvalue\"}]}}",
					Annotations: "{\"annotation-name-1\":\"annotation-value-1\",\"annotation-name-2\":\"annotation-value-2\"}",
					Labels:      "",
				},
			},
			expected: fakeAppId,
			wantErr:  false,
		},
		{
			name: "Test case with valid JSON labels",
			args: args{
				spec: &types.ServiceSpec{
					Name:        fakeAppName,
					Image:       "busybox",
					Cmd:         "sh",
					Args:        "-c \"sleep 1000\"",
					Instances:   1,
					CPU:         1,
					Mcpu:        1,
					Mem:         128,
					Ports:       []int{8080},
					Envs:        "[{\"name\":\"testkey\",\"value\":\"testValue\"}]",
					Liveness:    "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"tcpSocket\":{\"host\":\"\", \"port\":8080}}",
					Readiness:   "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"httpGet\": {\"path\": \"/consumer\",\"port\": 8080,\"scheme\": \"HTTP\",\"httpHeaders\": [{\"name\": \"test\",\"value\": \"testvalue\"}]}}",
					Annotations: "{\"annotation-name-1\":\"annotation-value-1\",\"annotation-name-2\":\"annotation-value-2\"}",
					Labels:      "{\"version\":\"1.0.0\",\"tier\":\"backend\"}",
				},
			},
			expected: fakeAppId,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.InsertK8sApp(tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("InsertK8sApp error, got: %v", err)
			}
			if got != tt.expected {
				t.Fatalf("InsertK8sApp error, got: %v, want: %v", got, tt.expected)
			}
		})
	}
}

func TestDeployK8sApplicationWithLabels(t *testing.T) {
	c, patches := newTestWrapEDASClient()
	defer patches.Reset()

	type args struct {
		appId string
		spec  *types.ServiceSpec
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test successful case with labels",
			args: args{
				appId: fakeAppId,
				spec:  fakeServiceSpec,
			},
			wantErr: false,
		},
		{
			name: "Test case with empty labels",
			args: args{
				appId: fakeAppId,
				spec: &types.ServiceSpec{
					Name:        "service-a",
					Image:       "busybox",
					Cmd:         "sh",
					Args:        "-c \"sleep 1000\"",
					Instances:   1,
					CPU:         1,
					Mcpu:        1,
					Mem:         128,
					Ports:       []int{8080},
					Envs:        "[{\"name\":\"testkey\",\"value\":\"testValue\"}]",
					Liveness:    "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"tcpSocket\":{\"host\":\"\", \"port\":8080}}",
					Readiness:   "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"httpGet\": {\"path\": \"/consumer\",\"port\": 8080,\"scheme\": \"HTTP\",\"httpHeaders\": [{\"name\": \"test\",\"value\": \"testvalue\"}]}}",
					Annotations: "{\"annotation-name-1\":\"annotation-value-1\",\"annotation-name-2\":\"annotation-value-2\"}",
					Labels:      "",
				},
			},
			wantErr: false,
		},
		{
			name: "Test case with complex labels",
			args: args{
				appId: fakeAppId,
				spec: &types.ServiceSpec{
					Name:        "service-a",
					Image:       "busybox",
					Cmd:         "sh",
					Args:        "-c \"sleep 1000\"",
					Instances:   1,
					CPU:         1,
					Mcpu:        1,
					Mem:         128,
					Ports:       []int{8080},
					Envs:        "[{\"name\":\"testkey\",\"value\":\"testValue\"}]",
					Liveness:    "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"tcpSocket\":{\"host\":\"\", \"port\":8080}}",
					Readiness:   "{\"failureThreshold\": 3,\"initialDelaySeconds\": 5,\"successThreshold\": 1,\"timeoutSeconds\": 1,\"httpGet\": {\"path\": \"/consumer\",\"port\": 8080,\"scheme\": \"HTTP\",\"httpHeaders\": [{\"name\": \"test\",\"value\": \"testvalue\"}]}}",
					Annotations: "{\"annotation-name-1\":\"annotation-value-1\",\"annotation-name-2\":\"annotation-value-2\"}",
					Labels:      "{\"app\":\"service-a\",\"servicegroup-id\":\"sg-123\",\"version\":\"v1.2.3\",\"environment\":\"production\"}",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.DeployApp(tt.args.appId, tt.args.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeployApp error, got: %v", err)
			}
		})
	}
}
