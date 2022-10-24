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

package v1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/ingress/common"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

const (
	projectNamespace = "project-ut-dev"
)

func Test_CreateIfNotExists(t *testing.T) {
	type args struct {
		client v1beta1.ExtensionsV1beta1Interface
		svc    *apistructs.Service
	}

	tests := []struct {
		name    string
		args    args
		created bool
		wantErr bool
	}{
		{
			name: "create",
			args: args{
				client: func() v1beta1.ExtensionsV1beta1Interface {
					return fakeclientset.NewSimpleClientset().ExtensionsV1beta1()
				}(),
				svc: &apistructs.Service{
					Name: "web",
					Labels: map[string]string{
						common.LabelHAProxyVHost: "localhost,ut.erda.cloud",
					},
					Namespace: projectNamespace,
					Ports: []diceyml.ServicePort{
						{
							Port: 8080,
						},
					},
				},
			},
			created: true,
		},
		{
			name: "service nil",
			args: args{
				client: func() v1beta1.ExtensionsV1beta1Interface {
					return fakeclientset.NewSimpleClientset().ExtensionsV1beta1()
				}(),
				svc: nil,
			},
			wantErr: true,
		},
		{
			name: "ingress existed",
			args: args{
				client: func() v1beta1.ExtensionsV1beta1Interface {
					return fakeclientset.NewSimpleClientset(&extensionsv1beta1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "web",
							Namespace: projectNamespace,
						},
					}).ExtensionsV1beta1()
				}(),
				svc: &apistructs.Service{
					Name: "web",
					Labels: map[string]string{
						common.LabelHAProxyVHost: "localhost,ut.erda.cloud",
					},
					Namespace: projectNamespace,
					Ports: []diceyml.ServicePort{
						{
							Port: 8080,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "none need ingress",
			args: args{
				client: func() v1beta1.ExtensionsV1beta1Interface {
					return fakeclientset.NewSimpleClientset().ExtensionsV1beta1()
				}(),
				svc: &apistructs.Service{
					Name:      "web",
					Namespace: projectNamespace,
					Ports: []diceyml.ServicePort{
						{
							Port: 8080,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ing := NewIngress(tt.args.client)
			err := ing.CreateIfNotExists(tt.args.svc)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_, err = tt.args.client.Ingresses(projectNamespace).Get(context.Background(), "web", metav1.GetOptions{})
			if tt.created && !k8serrors.IsNotFound(err) {
				assert.NoError(t, err)
			}
		})
	}
}
