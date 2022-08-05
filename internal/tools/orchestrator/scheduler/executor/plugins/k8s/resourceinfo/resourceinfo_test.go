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

package resourceinfo

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSplitPodsByNodeName(t *testing.T) {
	type args struct {
		podLi *corev1.PodList
	}

	pod1 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-1",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			NodeName: "node1",
		},
	}
	pod2 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-2",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			NodeName: "node2",
		},
	}
	pod3 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-3",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			NodeName: "node1",
		},
	}

	tests := []struct {
		name string
		args args
		want map[string][]corev1.Pod
	}{
		{
			name: "test1",
			args: args{
				podLi: &corev1.PodList{
					Items: []corev1.Pod{
						pod1, pod2, pod3,
					},
				},
			},
			want: map[string][]corev1.Pod{
				"node1": {pod1, pod3},
				"node2": {pod2},
			},
		},
		{
			name: "test2",
			args: args{
				podLi: &corev1.PodList{
					Items: []corev1.Pod{
						pod1, pod2,
					},
				},
			},
			want: map[string][]corev1.Pod{
				"node1": {pod1},
				"node2": {pod2},
			},
		},
		{
			name: "empty pod list",
			args: args{
				podLi: &corev1.PodList{
					Items: nil,
				},
			},
			want: map[string][]corev1.Pod{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitPodsByNodeName(tt.args.podLi); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitPodsByNodeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
