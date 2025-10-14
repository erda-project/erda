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
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sservice"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestCreateOrPutService(t *testing.T) {
	k8s := &Kubernetes{
		service: k8sservice.New(),
	}

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(k8s.service), "Get", func(*k8sservice.Service, string, string) (*apiv1.Service, error) {
		return &apiv1.Service{}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(k8s), "UpdateK8sService", func(*Kubernetes, *v1.Service, *apistructs.Service, map[string]string) error {
		return nil
	})

	err := k8s.CreateOrPutService(&apistructs.Service{
		Name:      "fake-service",
		Namespace: apiv1.NamespaceDefault,
		Ports: []diceyml.ServicePort{
			{Port: 80, Protocol: "TCP"},
		},
	}, map[string]string{})
	assert.NoError(t, err)
}

func Test_newService(t *testing.T) {
	type args struct {
		service *apistructs.Service
		labels  map[string]string
	}
	tests := []struct {
		name string
		args args
		want *v1.Service
	}{
		{
			name: "test new service",
			args: args{
				service: &apistructs.Service{
					Name:      "fake-service",
					Namespace: apiv1.NamespaceDefault,
					Ports: []diceyml.ServicePort{
						{Port: 80, Protocol: "TCP"},
					},
					Labels: map[string]string{
						"app":                    "fake-service",
						"mcp.erda.cloud/name":    "fake-service",
						"mcp.erda.cloud/version": "1.0.0",
						"DICE_ORG_ID":            "6",
					},
					Annotations: map[string]string{
						"mcp.erda.cloud/description": "This is a fake mcp server",
					},
				},
				labels: map[string]string{
					"servicegroup-id": "1",
					"app":             "fake-service",
					"svc":             "fake-service.default.svc.cluster.local",
					// invalid label, value must be 63 characters or less and must be empty or begin and end with an alphanumeric character ([a-z0-9A-Z])
					"invalid": "manager.addon-idxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.erda.cloud",
				},
			},
			want: &v1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-service",
					Namespace: apiv1.NamespaceDefault,
					Labels: map[string]string{
						"servicegroup-id":           "1",
						"app":                       "fake-service",
						"svc":                       "fake-service.default.svc.cluster.local",
						"mcp.erda.cloud/component":  "mcp-server",
						"mcp.erda.cloud/name":       "fake-service",
						"mcp.erda.cloud/version":    "1.0.0",
						"mcp.erda.cloud/scope-type": "org",
						"mcp.erda.cloud/scope-id":   "6",
					},
					Annotations: map[string]string{
						"mcp.erda.cloud/description": "This is a fake mcp server",
					},
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:       "tcp-0",
							Port:       80,
							TargetPort: intstr.FromInt(80),
						},
					},
					Selector: map[string]string{
						"app":             "fake-service",
						"servicegroup-id": "1",
						"svc":             "fake-service.default.svc.cluster.local",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newService(tt.args.service, tt.args.labels)
			if d := cmp.Diff(got, tt.want); d != "" {
				t.Errorf("newService() mismatch (-want +got):\n%s", d)
			}
		})
	}
}
