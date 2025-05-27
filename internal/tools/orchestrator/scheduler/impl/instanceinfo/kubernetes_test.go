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

package instanceinfo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/apistructs"
)

func Test_filterContainers(t *testing.T) {
	cases := []struct {
		name       string
		containers []corev1.Container
		expectLen  int
		expectName string
		expectVal  string
	}{
		{
			name:       "no containers",
			containers: nil,
			expectLen:  0,
			expectName: "",
			expectVal:  "",
		},
		{
			name:       "no DICE_CLUSTER_NAME",
			containers: []corev1.Container{{Name: "a"}},
			expectLen:  0,
			expectName: "",
			expectVal:  "",
		},
		{
			name: "one with DICE_CLUSTER_NAME",
			containers: []corev1.Container{{
				Name: "a",
				Env:  []corev1.EnvVar{{Name: apistructs.DICE_CLUSTER_NAME.String(), Value: "c1"}},
			}},
			expectLen:  1,
			expectName: "a",
			expectVal:  "c1",
		},
		{
			name: "multiple, only one valid",
			containers: []corev1.Container{{Name: "a"}, {
				Name: "b",
				Env:  []corev1.EnvVar{{Name: apistructs.DICE_CLUSTER_NAME.String(), Value: "c2"}},
			}},
			expectLen:  1,
			expectName: "b",
			expectVal:  "c2",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := filterContainers(c.containers)
			assert.Equal(t, c.expectLen, len(res.FilterContainers))
			if c.expectLen > 0 {
				assert.Equal(t, c.expectName, res.FilterContainers[0].Name)
				assert.Equal(t, c.expectVal, res.ClusterName)
			} else {
				assert.Equal(t, "", res.ClusterName)
			}
		})
	}
}

func Test_extractK8sPodsFromServiceGroup(t *testing.T) {
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p1"}}
	b, _ := json.Marshal([]corev1.Pod{pod})
	cases := []struct {
		name    string
		sg      *apistructs.ServiceGroup
		expects int
		hasErr  bool
	}{
		{
			name:    "nil sg",
			sg:      nil,
			expects: 0,
			hasErr:  true,
		},
		{
			name:    "nil extra",
			sg:      &apistructs.ServiceGroup{},
			expects: 0,
			hasErr:  true,
		},
		{
			name:    "no key",
			sg:      &apistructs.ServiceGroup{Extra: map[string]string{}},
			expects: 0,
			hasErr:  true,
		},
		{
			name:    "bad json",
			sg:      &apistructs.ServiceGroup{Extra: map[string]string{"svc": "{"}},
			expects: 0,
			hasErr:  true,
		},
		{
			name:    "ok",
			sg:      &apistructs.ServiceGroup{Extra: map[string]string{"svc": string(b)}},
			expects: 1,
			hasErr:  false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pods, err := extractK8sPodsFromServiceGroup(c.sg, "svc")
			if c.hasErr {
				assert.Error(t, err)
				assert.Nil(t, pods)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expects, len(pods))
			}
		})
	}
}

func Test_extractContainerID(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"docker://abc123", "abc123"},
		{"containerd://def456", "def456"},
		{"no-scheme", ""},
		{"", ""},
	}
	for _, c := range cases {
		assert.Equal(t, c.out, extractContainerID(c.in))
	}
}

func Test_getPodDefaultMessage(t *testing.T) {
	pod := corev1.Pod{Status: corev1.PodStatus{Message: "msg"}}
	assert.Equal(t, "msg", getPodDefaultMessage(pod))
	pod.Status.Message = ""
	assert.Equal(t, PodDefaultMessage, getPodDefaultMessage(pod))
}

func Test_getPodUpdateTime(t *testing.T) {
	now := metav1.Now()
	pod := corev1.Pod{
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, LastTransitionTime: now}},
			StartTime:  &now,
		},
		ObjectMeta: metav1.ObjectMeta{CreationTimestamp: now},
	}
	assert.Equal(t, now, *getPodUpdateTime(pod))

	pod.Status.Conditions = nil
	assert.Equal(t, now, *getPodUpdateTime(pod))

	pod.Status.StartTime = nil
	assert.Equal(t, now, *getPodUpdateTime(pod))
}

func Test_buildPodInfo(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-pod",
			Namespace:         "default",
			UID:               "uid-1",
			CreationTimestamp: metav1.Now(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "main",
				Env:  []corev1.EnvVar{{Name: apistructs.DICE_CLUSTER_NAME.String(), Value: "cluster-1"}},
			}},
		},
		Status: corev1.PodStatus{
			Phase:     corev1.PodRunning,
			PodIP:     "1.2.3.4",
			HostIP:    "2.3.4.5",
			StartTime: &metav1.Time{Time: time.Now()},
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:        "main",
				Ready:       true,
				ContainerID: "docker://cid-1",
				Image:       "busybox",
			}},
		},
	}
	podInfo, err := buildPodInfo(pod, "svc")
	assert.NoError(t, err)
	assert.Equal(t, "uid-1", podInfo.Uid)
	assert.Equal(t, "1.2.3.4", podInfo.IPAddress)
	assert.Equal(t, "cluster-1", podInfo.ClusterName)
	assert.Equal(t, "main", podInfo.PodContainers[0].ContainerName)
}

func Test_buildContainersResourcePending(t *testing.T) {
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "main",
				Image: "busybox",
				Env:   []corev1.EnvVar{{Name: apistructs.DICE_CLUSTER_NAME.String(), Value: "c1"}},
			}},
		},
		Status: corev1.PodStatus{},
	}
	msg := ""
	containers := buildContainersResourcePending(pod, &msg)
	assert.Equal(t, 1, len(containers))
	assert.Equal(t, "main", containers[0].ContainerName)
	assert.Equal(t, "busybox", containers[0].Image)
}

func Test_buildContainersResourceRunning(t *testing.T) {
	now := metav1.Now()
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:               "uid-1",
			CreationTimestamp: now,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "main",
				Image: "busybox",
				Env:   []corev1.EnvVar{{Name: apistructs.DICE_CLUSTER_NAME.String(), Value: "c1"}},
			}},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name:        "main",
				Ready:       true,
				ContainerID: "docker://cid-1",
				Image:       "busybox",
			}},
			PodIP:     "1.2.3.4",
			HostIP:    "2.3.4.5",
			StartTime: &now,
		},
	}
	status, containers, restart, msg, err := buildContainersResourceRunning(pod, "svc")
	assert.NoError(t, err)
	assert.Equal(t, PodStatusHealthy, status)
	assert.Equal(t, 1, len(containers))
	assert.Equal(t, "main", containers[0].ContainerName)
	assert.Equal(t, int32(0), restart)
	assert.NotEmpty(t, msg)
}

func Test_sortContainers(t *testing.T) {
	containers := []apistructs.PodContainer{{ContainerName: "sidecar"}, {ContainerName: "main"}}
	res := sortContainers(containers, "main")
	assert.Equal(t, "main", res[0].ContainerName)
}

func Test_getMainContainerNameByServiceName(t *testing.T) {
	containers := []corev1.Container{{
		Name: "main",
		Env:  []corev1.EnvVar{{Name: apistructs.EnvDiceServiceName, Value: "svc"}},
	}, {
		Name: "sidecar",
	}}
	name := getMainContainerNameByServiceName(containers, "svc")
	assert.Equal(t, "main", name)

	name2 := getMainContainerNameByServiceName(containers, "other")
	assert.Equal(t, "main", name2)
}
