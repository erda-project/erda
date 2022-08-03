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

package pod

import (
	"testing"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8serror"
	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

func TestPod_GetNamespacedPodsStatus(t *testing.T) {
	pod := Pod{}
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
			},
			Spec: corev1.PodSpec{},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				Conditions: []corev1.PodCondition{
					{
						Type:               corev1.PodInitialized,
						Status:             "True",
						LastTransitionTime: metav1.Time{},
					},
					{
						Type:               corev1.ContainersReady,
						Status:             "False",
						LastTransitionTime: metav1.Time{},
						Reason:             "ContainersNotReady",
						Message:            "containers with unready status: [test-pod]",
					},
					{
						Type:               corev1.PodScheduled,
						Status:             "True",
						LastTransitionTime: metav1.Time{},
					},
				},
				Message: "containers with unready status: [test-pod]",
				Reason:  "ContainersNotReady",
				HostIP:  "10.1.1.1",
				PodIP:   "9.8.7.6",
				PodIPs: []corev1.PodIP{
					{
						IP: "9.8.7.6",
					},
				},
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Image: "test-pod:v1",
						Name:  "test-pod",
						State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{
								Reason:  "ImagePullBackOff",
								Message: "Back-off pulling image \"test:v1\"",
							},
						},
					},
				},
				StartTime: &metav1.Time{},
			},
		},
	}
	podStatus, err := pod.GetNamespacedPodsStatus(pods, "")
	if err != nil {
		logrus.Fatal(err)
	}
	assert.DeepEqual(t, podStatus, []PodStatus{})

	podStatus, err = pod.GetNamespacedPodsStatus(pods, "test-pod")
	if err != nil {
		logrus.Fatal(err)
	}
	assert.DeepEqual(t, podStatus, []PodStatus{
		{
			Reason:  ImagePullFailed,
			Message: "Back-off pulling image \"test:v1\"",
		},
	})
}

func TestListAllNamespace(t *testing.T) {
	p := New(WithK8sClient(fakeclientset.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"core.erda.cloud/component": "orchestrator",
			},
			Namespace: metav1.NamespaceDefault,
		},
	})))
	pods, err := p.ListAllNamespace([]string{"core.erda.cloud/component=orchestrator"})
	assert.NilError(t, err)
	assert.Equal(t, len(pods.Items), 1)

	_, err = p.ListNamespacePods(metav1.NamespaceDefault)
	assert.NilError(t, err)
	assert.Equal(t, len(pods.Items), 1)

	err = p.Delete(metav1.NamespaceDefault, "test-pod")
	assert.NilError(t, err)

	_, err = p.Get(metav1.NamespaceDefault, "test-pod")
	assert.Equal(t, err, k8serror.ErrNotFound)
}
