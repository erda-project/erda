// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package pod

import (
	"testing"

	"github.com/sirupsen/logrus"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
