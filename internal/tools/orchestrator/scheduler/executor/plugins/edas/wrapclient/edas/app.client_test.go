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
}

func mock(c *api.Client) *gomonkey.Patches {
	patches := gomonkey.NewPatches()

	patches.ApplyMethod(reflect.TypeOf(c), "ListApplication", func(c *api.Client,
		req *api.ListApplicationRequest) (*api.ListApplicationResponse, error) {
		var apps []api.ApplicationInListApplication

		if req.AppName == fakeAppName {
			apps = append(apps, api.ApplicationInListApplication{
				AppId: fakeAppId,
				Name:  fakeAppName,
			})
		}
		return &api.ListApplicationResponse{
			RequestId: uuid.New(),
			ApplicationList: api.ApplicationList{
				Application: apps,
			},
		}, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "StopK8sApplication", func(c *api.Client,
		req *api.StopK8sApplicationRequest) (*api.StopK8sApplicationResponse, error) {
		resp := &api.StopK8sApplicationResponse{
			RequestId: uuid.New(),
		}

		if req.AppId == fakeAppId {
			resp.ChangeOrderId = fakeChangeOrderId
		}

		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "ListRecentChangeOrder", func(c *api.Client,
		req *api.ListRecentChangeOrderRequest) (*api.ListRecentChangeOrderResponse, error) {
		resp := &api.ListRecentChangeOrderResponse{
			RequestId: uuid.New(),
		}

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

		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "AbortChangeOrder", func(c *api.Client,
		req *api.AbortChangeOrderRequest) (*api.AbortChangeOrderResponse, error) {
		resp := &api.AbortChangeOrderResponse{
			RequestId: uuid.New(),
		}
		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "DeleteK8sApplication", func(c *api.Client,
		req *api.DeleteK8sApplicationRequest) (*api.DeleteK8sApplicationResponse, error) {
		resp := &api.DeleteK8sApplicationResponse{
			RequestId: uuid.New(),
		}
		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "GetChangeOrderInfo", func(c *api.Client,
		req *api.GetChangeOrderInfoRequest) (*api.GetChangeOrderInfoResponse, error) {
		resp := &api.GetChangeOrderInfoResponse{
			RequestId: uuid.New(),
		}

		if req.ChangeOrderId == fakeChangeOrderId {
			resp.ChangeOrderInfo = api.ChangeOrderInfo{
				ChangeOrderId: fakeChangeOrderId,
				Status:        int(types.CHANGE_ORDER_STATUS_SUCC),
			}
		}

		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "GetAppDeployment", func(c *api.Client,
		req *api.GetAppDeploymentRequest) (*api.GetAppDeploymentResponse, error) {
		resp := &api.GetAppDeploymentResponse{
			RequestId: uuid.New(),
		}

		if req.AppId == fakeAppId {
			deployJson, err := json.Marshal(fakeDeployment)
			if err != nil {
				return nil, err
			}
			resp.Data = string(deployJson)
		}

		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "ScaleK8sApplication", func(c *api.Client,
		req *api.ScaleK8sApplicationRequest) (*api.ScaleK8sApplicationResponse, error) {
		resp := &api.ScaleK8sApplicationResponse{
			RequestId: uuid.New(),
		}

		if req.AppId != fakeAppId {
			return nil, errors.New("scale fail")
		}

		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "DeployK8sApplication", func(c *api.Client,
		req *api.DeployK8sApplicationRequest) (*api.DeployK8sApplicationResponse, error) {
		resp := &api.DeployK8sApplicationResponse{
			RequestId:     uuid.New(),
			ChangeOrderId: fakeChangeOrderId,
		}

		if req.AppId != fakeAppId {
			return nil, errors.New("deploy fail")
		}

		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "InsertK8sApplication", func(c *api.Client,
		req *api.InsertK8sApplicationRequest) (*api.InsertK8sApplicationResponse, error) {
		resp := api.CreateInsertK8sApplicationResponse()
		resp.RequestId = uuid.New()
		resp.ApplicationInfo = api.ApplicationInfo{
			AppId:         fakeAppId,
			ChangeOrderId: fakeChangeOrderId,
		}

		if req.AppName != fakeAppName {
			return nil, errors.New("insert k8s application fail")
		}

		return resp, nil
	})

	patches.ApplyMethod(reflect.TypeOf(c), "QueryApplicationStatus", func(c *api.Client,
		req *api.QueryApplicationStatusRequest) (*api.QueryApplicationStatusResponse, error) {
		resp := api.CreateQueryApplicationStatusResponse()
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

		if req.AppId != fakeAppId {
			return nil, errors.New("query application fail")
		}

		return resp, nil
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
