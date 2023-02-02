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
	"context"
	"github.com/erda-project/erda/apistructs"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"net/http"
	"reflect"
	"testing"

	"bou.ke/monkey"
	api "github.com/aliyun/alibaba-cloud-sdk-go/services/edas"
)

func Test_Set_Annotations(t *testing.T) {
	type args struct {
		name    string
		envs    map[string]string
		svcSpec *ServiceSpec
		wantErr bool
	}

	svc := &ServiceSpec{
		Name:  "fake-app",
		Image: "busybox",
	}

	tests := []args{
		{
			name:    "spec nil",
			wantErr: true,
		},
		{
			name:    "annotations",
			svcSpec: svc,
			envs: map[string]string{
				"DICE_ORG_ID":     "org-id",
				"DICE_RUNTIME_ID": "runtime-id",
			},
		},
		{
			name:    "empty annotations",
			svcSpec: svc,
			envs: map[string]string{
				"FAKE_KEY": "fake-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setAnnotations(tt.svcSpec, tt.envs); (err != nil) != tt.wantErr {
				t.Errorf("SetAnnotations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func initClient() *api.Client {
	c := &api.Client{}
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "StopK8sApplication", func(c *api.Client,
		request *api.StopK8sApplicationRequest) (response *api.StopK8sApplicationResponse, err error) {
		return &api.StopK8sApplicationResponse{
			Code: http.StatusOK,
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "DeleteK8sApplication", func(c *api.Client,
		request *api.DeleteK8sApplicationRequest) (response *api.DeleteK8sApplicationResponse, err error) {
		return &api.DeleteK8sApplicationResponse{
			Code: http.StatusOK,
		}, nil
	})
	return c
}

func TestDeleteAppByID(t *testing.T) {
	c := initClient()
	e := EDAS{
		addr:   "cn-hangzhou",
		client: c,
	}

	defer monkey.UnpatchAll()

	if err := e.deleteAppByID("app1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStopAppByID(t *testing.T) {
	c := initClient()
	e := EDAS{
		addr:   "cn-hangzhou",
		client: c,
	}

	defer monkey.UnpatchAll()

	if err := e.stopAppByID("app1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatus(t *testing.T) {
	fc := fakeclientset.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-1-app-demo",
			Namespace: defaultNamespace,
			Labels: map[string]string{
				"edas-domain":  "edas-admin",
				"edas.appname": "service-1-app-demo",
			},
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{
				{
					Type:   appsv1.DeploymentAvailable,
					Status: "True",
				},
			},
			Replicas:          1,
			ReadyReplicas:     1,
			AvailableReplicas: 1,
			UpdatedReplicas:   1,
		},
	})

	e := EDAS{
		cs: fc,
	}

	type args struct {
		specObject interface{}
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				specObject: apistructs.ServiceGroup{
					Dice: apistructs.Dice{
						ID:   "1",
						Type: "service",
						Services: []apistructs.Service{
							{
								Name:  "app-demo",
								Image: "busybox",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: add more ut
			_, err := e.Status(context.Background(), tt.args.specObject)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
