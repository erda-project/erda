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

package deployment

import (
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
	"k8s.io/utils/pointer"
)

func Test_Patch(t *testing.T) {
	// prepare data
	cs := fakeclientset.NewSimpleClientset()
	d := New(WithClientSet(cs))

	type args struct {
		name          string
		namespace     string
		containerName string
		patch         *diceyml.K8SSnippet
	}

	percent := intstr.FromString("20%")
	var minReadySeconds int32 = 60
	tests := []struct {
		name string
		args args
		want func(*appsv1.Deployment)
	}{
		{
			name: "case 1",
			args: args{
				name:          "hello",
				namespace:     "test-namespace",
				containerName: "container-1",
				patch: &diceyml.K8SSnippet{
					Container: &diceyml.ContainerSnippet{
						ImagePullPolicy: corev1.PullNever,
						SecurityContext: &corev1.SecurityContext{
							Privileged: pointer.Bool(true),
						},
					},
				},
			},
			want: func(d *appsv1.Deployment) {
				d.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullNever
				d.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
					Privileged: pointer.Bool(true),
				}
			},
		}, {
			name: "case 2",
			args: args{
				name:          "hello",
				namespace:     "test-namespace",
				containerName: "container-2",
				patch: &diceyml.K8SSnippet{
					Workload: &diceyml.WorkloadSnippet{
						Deployment: &diceyml.DeploymentSnippet{
							MinReadySeconds: &minReadySeconds,
							Strategy: &appsv1.DeploymentStrategy{
								Type: appsv1.RollingUpdateDeploymentStrategyType,
								RollingUpdate: &appsv1.RollingUpdateDeployment{
									MaxUnavailable: &percent,
									MaxSurge:       &percent,
								},
							},
						},
					},
				},
			},
			want: func(d *appsv1.Deployment) {
				d.Spec.MinReadySeconds = 60
				d.Spec.Strategy = appsv1.DeploymentStrategy{
					Type: appsv1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &appsv1.RollingUpdateDeployment{
						MaxUnavailable: &percent,
						MaxSurge:       &percent,
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploy := genDeployment(tt.args.name, tt.args.namespace)
			if err := d.Create(deploy); err != nil {
				t.Fatal(err)
			}

			if err := d.Patch(tt.args.namespace, tt.args.name,
				tt.args.containerName, tt.args.patch); err != nil {
				t.Fatal(err)
			}

			patchedDeploy, err := d.Get(tt.args.namespace, tt.args.name)
			if err != nil {
				t.Fatal(err)
			}

			tt.want(deploy)

			if !reflect.DeepEqual(patchedDeploy, deploy) {
				t.Errorf("Patch(), got %v, want %v", patchedDeploy, deploy)
			}

			if err = d.Delete(tt.args.namespace, tt.args.name); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_List(t *testing.T) {
	deployList := []appsv1.Deployment{
		*genDeployment("demo1", metav1.NamespaceDefault, "target=select"),
		*genDeployment("demo2", metav1.NamespaceDefault),
		*genDeployment("demo3", metav1.NamespaceDefault),
		*genDeployment("demo4", metav1.NamespaceDefault, "target=select"),
		*genDeployment("demo5", "non-system-default"),
	}
	// prepare data
	cs := &fakeclientset.Clientset{}
	cs.AddReactor("*", "deployments", func(action clientgotesting.Action) (handled bool,
		ret runtime.Object, err error) {
		return true, &appsv1.DeploymentList{
			ListMeta: metav1.ListMeta{
				ResourceVersion: "1",
			},
			Items: deployList,
		}, nil
	})
	d := New(WithClientSet(cs))

	deploys, err := d.List(metav1.NamespaceDefault, map[string]string{
		"target": "select",
	})
	if err != nil {
		t.Fatal(err)
	}

	got := len(deploys.Items)
	if got != 2 {
		t.Fatalf("List with label seletor error, got %v, want 2", got)
	}

	// fake clientSet can't use fieldSelector filter
	_, _, err = d.LimitedListAllNamespace(100, nil)
}

func genDeployment(name, namespace string, label ...string) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "busybox",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "busybox",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-1",
							Image: "busybox",
						},
					},
				},
			},
		},
	}

	if len(label) != 0 {
		labels := make(map[string]string)
		for _, l := range label {
			kv := strings.Split(l, "=")
			if len(kv) != 2 {
				continue
			}
			labels[kv[0]] = kv[1]
		}
		deploy.Labels = labels
	}

	return deploy
}
