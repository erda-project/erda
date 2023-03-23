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
	"testing"

	corev1 "k8s.io/api/core/v1"

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
		want bool
	}{
		{
			name: "production environment",
			args: args{
				service: &apistructs.Service{
					Env: map[string]string{
						apistructs.DiceWorkspaceEnvKey: apistructs.ProdWorkspace.String(),
					},
				},
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: "busybox",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "none production environment",
			args: args{
				service: &apistructs.Service{
					Env: map[string]string{
						apistructs.DiceWorkspaceEnvKey: apistructs.TestWorkspace.String(),
					},
				},
				podSpec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: "busybox",
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k.AddLifeCycle(tt.args.service, tt.args.podSpec)
			got := tt.args.podSpec.Containers[0].Lifecycle != nil &&
				tt.args.podSpec.TerminationGracePeriodSeconds != nil
			if got != tt.want {
				t.Fatalf("add lifecycle fail, got: lifecycle: %+v, termination grace period seconds: %v",
					tt.args.podSpec.Containers[0].Lifecycle, tt.args.podSpec.TerminationGracePeriodSeconds)
			}
		})
	}
}
