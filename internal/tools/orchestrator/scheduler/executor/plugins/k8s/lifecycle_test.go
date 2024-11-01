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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/erda-project/erda/apistructs"
)

func TestAddLifecycle(t *testing.T) {
	k := Kubernetes{}

	type args struct {
		service *apistructs.Service
		podSpec *corev1.PodSpec
	}
	tests := []struct {
		name string
		args args
		want *corev1.PodSpec
	}{
		{
			name: "nil podSpec",
			args: args{
				service: &apistructs.Service{
					Env: map[string]string{
						apistructs.DiceWorkspaceEnvKey: apistructs.ProdWorkspace.String(),
					},
				},
				podSpec: nil,
			},
			want: nil,
		},
		{
			name: "empty containers",
			args: args{
				service: &apistructs.Service{
					Env: map[string]string{
						apistructs.DiceWorkspaceEnvKey: apistructs.ProdWorkspace.String(),
					},
				},
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
			want: &corev1.PodSpec{
				Containers: []corev1.Container{},
			},
		},
		{
			name: "non-PROD workspace",
			args: args{
				service: &apistructs.Service{
					Env: map[string]string{
						apistructs.DiceWorkspaceEnvKey: "DEV",
					},
				},
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
			want: &corev1.PodSpec{
				Containers: []corev1.Container{{}},
			},
		},
		{
			name: "PROD workspace with containers",
			args: args{
				service: &apistructs.Service{
					Env: map[string]string{
						apistructs.DiceWorkspaceEnvKey: apistructs.ProdWorkspace.String(),
					},
				},
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
			want: &corev1.PodSpec{
				TerminationGracePeriodSeconds: pointer.Int64(DefaultProdTerminationGracePeriodSeconds),
				Containers: []corev1.Container{{
					Lifecycle: &corev1.Lifecycle{
						PreStop: DefaultProdLifecyclePreStopHandler,
					},
				}},
			},
		},
		{
			name: "container with existing Lifecycle and PostStart",
			args: args{
				service: &apistructs.Service{
					Env: map[string]string{
						apistructs.DiceWorkspaceEnvKey: apistructs.ProdWorkspace.String(),
					},
				},
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{{
						Lifecycle: &corev1.Lifecycle{
							PostStart: &corev1.LifecycleHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"echo", "postStart"},
								},
							},
						},
					}},
				},
			},
			want: &corev1.PodSpec{
				TerminationGracePeriodSeconds: pointer.Int64(DefaultProdTerminationGracePeriodSeconds),
				Containers: []corev1.Container{{
					Lifecycle: &corev1.Lifecycle{
						PostStart: &corev1.LifecycleHandler{
							Exec: &corev1.ExecAction{
								Command: []string{"echo", "postStart"},
							},
						},
						PreStop: DefaultProdLifecyclePreStopHandler,
					},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k.AddLifeCycle(tt.args.service, tt.args.podSpec)
			if !reflect.DeepEqual(tt.args.podSpec, tt.want) {
				t.Errorf("AddLifeCycle() = %v, want %v", tt.args.podSpec, tt.want)
			}
		})
	}
}
