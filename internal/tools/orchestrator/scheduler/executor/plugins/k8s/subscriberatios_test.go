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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sapi"
)

func Test_CalcFineGrainedCPU(t *testing.T) {
	k := &Kubernetes{}

	tests := []struct {
		name               string
		requestCPU         float64
		maxCPU             float64
		ratio              float64
		expectedRequestCPU float64
		expectedLimitCPU   float64
		expectedError      bool
	}{
		{
			name:               "valid input with maxCPU set",
			requestCPU:         1,
			maxCPU:             2,
			ratio:              10,
			expectedRequestCPU: 1,
			expectedLimitCPU:   2,
		},
		{
			name:               "valid input without maxCPU",
			requestCPU:         1,
			maxCPU:             0,
			ratio:              5,
			expectedRequestCPU: 0.2,
			expectedLimitCPU:   1,
		},
		{
			name:               "requestCPU less than MIN_CPU_SIZE",
			requestCPU:         0.05,
			maxCPU:             2.0,
			ratio:              1.5,
			expectedRequestCPU: 0,
			expectedLimitCPU:   0,
			expectedError:      true,
		},
		{
			name:               "maxCPU less than requestCPU",
			requestCPU:         2.0,
			maxCPU:             1.0,
			ratio:              1.5,
			expectedRequestCPU: 0,
			expectedLimitCPU:   0,
			expectedError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualCPU, actualMaxCPU, err := k.calcFineGrainedCPU(tt.requestCPU, tt.maxCPU, tt.ratio)
			if err != nil {
				if tt.expectedError {
					t.Logf("want error, got %v", err)
					return
				}
				t.Fatalf("want no error, got %v", err)
			}

			// Assert CPU values
			assert.Equal(t, tt.expectedRequestCPU, actualCPU)
			assert.Equal(t, tt.expectedLimitCPU, actualMaxCPU)
		})
	}
}

func TestCalcFineGrainedMemory(t *testing.T) {
	tests := []struct {
		name               string
		requestMem         float64
		maxMem             float64
		memSubscribeRatio  float64
		expectedRequestMem float64
		expectedMaxMem     float64
		expectedError      bool
	}{
		{
			name:               "Valid case with non-zero maxMem",
			requestMem:         512,
			maxMem:             1024,
			memSubscribeRatio:  10,
			expectedRequestMem: 512,
			expectedMaxMem:     1024,
		},
		{
			name:               "Valid case with zero maxMem",
			requestMem:         512,
			maxMem:             0,
			memSubscribeRatio:  10,
			expectedRequestMem: 51.2,
			expectedMaxMem:     512,
		},
		{
			name:              "Invalid requestMem less than MIN_MEM_SIZE",
			requestMem:        5,
			maxMem:            100,
			memSubscribeRatio: 10,
			expectedError:     true,
		},
		{
			name:              "Invalid maxMem less than requestMem",
			requestMem:        512,
			maxMem:            256,
			memSubscribeRatio: 10,
			expectedError:     true,
		},
		{
			name:               "Zero maxMem and no memSubscribeRatio",
			requestMem:         512,
			memSubscribeRatio:  1,
			expectedRequestMem: 512,
			expectedMaxMem:     512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kubernetes{}
			reqMem, maxMem, err := k.calcFineGrainedMemory(tt.requestMem, tt.maxMem, tt.memSubscribeRatio)
			if err != nil {
				if tt.expectedError {
					t.Logf("want error, got %v", err)
					return
				}
				t.Fatalf("want no error, got %v", err)
			}

			assert.Equal(t, tt.expectedRequestMem, reqMem)
			assert.Equal(t, tt.expectedMaxMem, maxMem)
		})
	}
}

func TestGetWorkspaceRatio(t *testing.T) {
	tests := []struct {
		name          string
		options       map[string]string
		workspace     string
		expectedValue float64
	}{
		{
			name:          "Default ratio when no options provided",
			options:       map[string]string{},
			expectedValue: DefaultRatio,
		},
		{
			name:          "Set ratio for production",
			options:       map[string]string{"CPU_SUBSCRIBE_RATIO": "10"},
			workspace:     apistructs.ProdWorkspace.String(),
			expectedValue: 10,
		},
		{
			name:          "Non-production workspace",
			options:       map[string]string{"DEV_CPU_SUBSCRIBE_RATIO": "10"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 10,
		},
		{
			name:          "Non-production workspace overrides global",
			options:       map[string]string{"CPU_SUBSCRIBE_RATIO": "10", "DEV_CPU_SUBSCRIBE_RATIO": "20"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 20,
		},
		{
			name:          "Ratio < 1.0",
			options:       map[string]string{"CPU_SUBSCRIBE_RATIO": "10", "DEV_CPU_SUBSCRIBE_RATIO": "0.5"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 10,
		},
		{
			name:          "Ratio < 1.0 case2",
			options:       map[string]string{"DEV_CPU_SUBSCRIBE_RATIO": "0.5"},
			workspace:     apistructs.DevWorkspace.String(),
			expectedValue: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value float64
			getWorkspaceRatio(tt.options, tt.workspace, "CPU", &value)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

func TestGetSubscribeRationsByWorkspace(t *testing.T) {
	k8s := &Kubernetes{
		devCpuSubscribeRatio:     10,
		devMemSubscribeRatio:     2,
		testCpuSubscribeRatio:    3,
		testMemSubscribeRatio:    5,
		stagingCpuSubscribeRatio: 20,
		stagingMemSubscribeRatio: 10,
		cpuSubscribeRatio:        10,
		memSubscribeRatio:        1,
	}

	tests := []struct {
		workspace        apistructs.DiceWorkspace
		expectedCPURatio float64
		expectedMemRatio float64
	}{
		{
			workspace:        apistructs.DevWorkspace,
			expectedCPURatio: 10,
			expectedMemRatio: 2,
		},
		{
			workspace:        apistructs.TestWorkspace,
			expectedCPURatio: 3,
			expectedMemRatio: 5,
		},
		{
			workspace:        apistructs.StagingWorkspace,
			expectedCPURatio: 20,
			expectedMemRatio: 10,
		},
		{
			workspace:        apistructs.ProdWorkspace,
			expectedCPURatio: 10,
			expectedMemRatio: 1,
		},
		{
			workspace:        apistructs.DiceWorkspace("UnknownWorkspace"),
			expectedCPURatio: DefaultRatio,
			expectedMemRatio: DefaultRatio,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.workspace), func(t *testing.T) {
			cpuRatio, memRatio := k8s.getSubscribeRationsByWorkspace(tt.workspace)
			assert.Equal(t, tt.expectedCPURatio, cpuRatio)
			assert.Equal(t, tt.expectedMemRatio, memRatio)
		})
	}
}

func TestResourceOverCommit(t *testing.T) {
	var tests = []struct {
		name                         string
		workspace                    apistructs.DiceWorkspace
		getSubscribeRation           func() *Kubernetes
		serviceResource              apistructs.Resources
		expectedResourceRequirements corev1.ResourceRequirements
	}{
		{
			name:      "dev workspace with resource over commit",
			workspace: apistructs.DevWorkspace,
			getSubscribeRation: func() *Kubernetes {
				return &Kubernetes{
					devCpuSubscribeRatio: 10,
					devMemSubscribeRatio: 1,
					cpuSubscribeRatio:    20,
					memSubscribeRatio:    1,
				}
			},
			serviceResource: apistructs.Resources{
				Cpu: 0.1,
				Mem: 1024,
			},
			expectedResourceRequirements: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("10m"),
					corev1.ResourceMemory:           resource.MustParse("1024Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("100m"),
					corev1.ResourceMemory:           resource.MustParse("1024Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeLimit),
				},
			},
		},
		// OVer commit setting:
		// CPUSubscribeRatio: 10, DEVCPUSubscribeRatio not set
		// DEVCPUSubscribeRatio -> 10, test cases in TestGetWorkspaceRatio
		//{
		//	name:      "dev workspace use default resource over commit",
		//	workspace: apistructs.DevWorkspace,
		//},
		{
			name:      "prod workspace use default resource over commit",
			workspace: apistructs.ProdWorkspace,
			getSubscribeRation: func() *Kubernetes {
				return &Kubernetes{
					cpuSubscribeRatio: 10,
					memSubscribeRatio: 10,
				}
			},
			serviceResource: apistructs.Resources{
				Cpu: 0.1,
				Mem: 1024,
			},
			expectedResourceRequirements: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("10m"),
					corev1.ResourceMemory:           resource.MustParse("102Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("100m"),
					corev1.ResourceMemory:           resource.MustParse("1024Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeLimit),
				},
			},
		},
		{
			name:      "none resource over commit",
			workspace: apistructs.DevWorkspace,
			getSubscribeRation: func() *Kubernetes {
				return &Kubernetes{
					cpuSubscribeRatio: 10,
					memSubscribeRatio: 10,
				}
			},
			serviceResource: apistructs.Resources{
				Cpu:    0.1,
				Mem:    1024,
				MaxCPU: 0.5,
				MaxMem: 2048,
			},
			expectedResourceRequirements: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("100m"),
					corev1.ResourceMemory:           resource.MustParse("1024Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("500m"),
					corev1.ResourceMemory:           resource.MustParse("2048Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeLimit),
				},
			},
		},
		{
			name:      "ephemeral storage size limit over commit",
			workspace: apistructs.ProdWorkspace,
			getSubscribeRation: func() *Kubernetes {
				return &Kubernetes{}
			},
			serviceResource: apistructs.Resources{
				Cpu:                      0.1,
				Mem:                      1024,
				MaxCPU:                   0.5,
				MaxMem:                   2048,
				EphemeralStorageCapacity: 1024,
			},
			expectedResourceRequirements: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("100m"),
					corev1.ResourceMemory:           resource.MustParse("1024Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("500m"),
					corev1.ResourceMemory:           resource.MustParse("2048Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse("1024Gi"),
				},
			},
		},
		{
			name:      "cause zero resources large over commit, resource set min",
			workspace: apistructs.DevWorkspace,
			getSubscribeRation: func() *Kubernetes {
				return &Kubernetes{
					devCpuSubscribeRatio: 10,
					devMemSubscribeRatio: 20,
				}
			},
			serviceResource: apistructs.Resources{
				Cpu: 0.1,
				Mem: 10,
			},
			expectedResourceRequirements: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("10m"),
					corev1.ResourceMemory:           resource.MustParse("10Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeRequest),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:              resource.MustParse("100m"),
					corev1.ResourceMemory:           resource.MustParse("10Mi"),
					corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeLimit),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.getSubscribeRation()
			resourceRequirements, err := k.ResourceOverCommit(tt.workspace, tt.serviceResource)
			assert.NoError(t, err)

			if !reflect.DeepEqual(tt.expectedResourceRequirements, resourceRequirements) {
				t.Fatalf("resource requirements are not equal, got: %v, expected: %v",
					resourceRequirements, tt.expectedResourceRequirements)
			}
		})
	}
}
