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
