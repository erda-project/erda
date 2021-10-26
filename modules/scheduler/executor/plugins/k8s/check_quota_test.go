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
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestResourceToString(t *testing.T) {
	cpu := 1000.0
	mem := float64(1 << 30)
	cpuStr := resourceToString(cpu, "cpu")
	memStr := resourceToString(mem, "memory")
	if cpuStr != "1" {
		t.Errorf("test failed, expected cpu is \"1\", got %s", cpuStr)
	}
	if memStr != "1G" {
		t.Errorf("test failed, expected cpu is \"1G\", got %s", memStr)
	}
}

func TestGetRequestResources(t *testing.T) {
	containers := []corev1.Container{
		{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
					corev1.ResourceMemory: *resource.NewQuantity(1<<30, resource.DecimalSI),
				},
			},
		},
		{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
					corev1.ResourceMemory: *resource.NewQuantity(100<<20, resource.BinarySI),
				},
			},
		},
	}
	cpu, mem := getRequestsResources(containers)
	if cpu != 1500 {
		t.Errorf("test failed, expected cpu is 1500, got %d", cpu)
	}
	if mem != 1178599424 {
		t.Errorf("test failed, expected mem is 1178599424, got %d", mem)
	}
}
