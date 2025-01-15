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
	"strconv"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
)

const (
	DefaultRatio = 1.0
)

type SubscribeRatios struct {
	CPURatio float64
	MemRatio float64
}

// getWorkspaceRatio
func getWorkspaceRatio(options map[string]string, workspace string, t string, value *float64) {
	// Default subscribe ratio
	*value = DefaultRatio

	f := func(workspace string) {
		ratioValue, ok := options[workspace]
		if !ok {
			return
		}
		ratio, err := strconv.ParseFloat(ratioValue, 64)
		if err == nil && ratio >= DefaultRatio {
			*value = ratio
		}
	}

	// Set ratio to Global&PROD
	f(t + SUBSCRIBE_RATIO_SUFFIX)
	// If workspace is production, return
	if workspace == apistructs.ProdWorkspace.String() {
		return
	}

	// If non-production workspace existed, overwrite.
	f(workspace + "_" + t + SUBSCRIBE_RATIO_SUFFIX)
}

func (k *Kubernetes) calcFineGrainedCPU(requestCPU, maxCPU, ratio float64) (float64, float64, error) {
	// 1, Processing request cpu value
	actualCPU := requestCPU

	if requestCPU < MIN_CPU_SIZE {
		return 0, 0, errors.Errorf("invalid request cpu, value: %v, (which is lower than min cpu(%v))",
			requestCPU, MIN_CPU_SIZE)
	}

	// max_cpu set but smaller than request cpu
	if maxCPU != 0 && maxCPU < requestCPU {
		return 0, 0, errors.Errorf("invalid max cpu, value: %v, (which is lower than request cpu(%v))", maxCPU, requestCPU)
	}

	// if max_cpu not set, use [cpu/ratio, cpu]; else use [cpu, max_cpu]
	if maxCPU == 0 {
		maxCPU = requestCPU
		actualCPU = requestCPU / ratio
	}

	return actualCPU, maxCPU, nil
}

func (k *Kubernetes) calcFineGrainedMemory(requestMem, maxMem, memSubscribeRatio float64) (float64, float64, error) {
	if requestMem < MIN_MEM_SIZE {
		return 0, 0, errors.Errorf("invalid request mem, value: %v, (which is lower than min mem(%vMi))",
			requestMem, MIN_MEM_SIZE)
	}

	// max_mem set but smaller than request mem
	if maxMem != 0 && maxMem < requestMem {
		return 0, 0, errors.Errorf("invalid max mem, value: %v, (which is lower than request mem(%v))", maxMem, requestMem)
	}

	// if max_mem not set, use [mem/ratio, mem]; else use [mem, max_mem]
	if maxMem == 0 {
		maxMem = requestMem
		requestMem = requestMem / memSubscribeRatio
	}

	return requestMem, maxMem, nil
}

func (k *Kubernetes) ResourceInfo(brief bool) (apistructs.ClusterResourceInfoData, error) {
	r, err := k.resourceInfo.Get(brief)
	if err != nil {
		return r, err
	}
	r.ProdCPUOverCommit = k.cpuSubscribeRatio
	r.DevCPUOverCommit = k.devCpuSubscribeRatio
	r.TestCPUOverCommit = k.testCpuSubscribeRatio
	r.StagingCPUOverCommit = k.stagingCpuSubscribeRatio
	r.ProdMEMOverCommit = k.memSubscribeRatio
	r.DevMEMOverCommit = k.devMemSubscribeRatio
	r.TestMEMOverCommit = k.testMemSubscribeRatio
	r.StagingMEMOverCommit = k.stagingMemSubscribeRatio

	return r, nil
}

// getSubscribeRationsByWorkspace
// Args: workspace
// Return: cpu subscribe ratio, memory subscribe ratio
func (k *Kubernetes) getSubscribeRationsByWorkspace(workspace apistructs.DiceWorkspace) (float64, float64) {
	subscribeRatios := map[apistructs.DiceWorkspace]SubscribeRatios{
		apistructs.DevWorkspace:     {CPURatio: k.devCpuSubscribeRatio, MemRatio: k.devMemSubscribeRatio},
		apistructs.TestWorkspace:    {CPURatio: k.testCpuSubscribeRatio, MemRatio: k.testMemSubscribeRatio},
		apistructs.StagingWorkspace: {CPURatio: k.stagingCpuSubscribeRatio, MemRatio: k.stagingMemSubscribeRatio},
		apistructs.ProdWorkspace:    {CPURatio: k.cpuSubscribeRatio, MemRatio: k.memSubscribeRatio},
	}

	subscribeRatio, ok := subscribeRatios[workspace]
	if !ok {
		return DefaultRatio, DefaultRatio
	}

	return subscribeRatio.CPURatio, subscribeRatio.MemRatio
}

func (k *Kubernetes) ResourceOverCommit(workspace apistructs.DiceWorkspace, r apistructs.Resources) (corev1.ResourceRequirements, error) {
	// If workspace is "", use default ratio -> 1;
	// Get subscribe rations by workspace
	cpuRatio, memRatio := k.getSubscribeRationsByWorkspace(workspace)

	requestCPU, limitCPU, err := k.calcFineGrainedCPU(r.Cpu, r.MaxCPU, cpuRatio)
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	requestMem, limitMem, err := k.calcFineGrainedMemory(r.Mem, r.MaxMem, memRatio)
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	maxEphemeral := resource.MustParse(k8sapi.EphemeralStorageSizeLimit)
	if r.EphemeralStorageCapacity > 1 {
		maxEphemeral = util.ResourceEphemeralStorageCapacityFormatter(r.EphemeralStorageCapacity)
	}

	// If calculated over commit resource is zero, set platform min resources.
	var (
		actualRequestCPU = int(1000 * requestCPU)
		actualRequestMem = int(requestMem)
	)

	if actualRequestCPU == 0 {
		actualRequestCPU = int(1000 * MIN_CPU_SIZE)
	}
	if actualRequestMem == 0 {
		actualRequestMem = MIN_MEM_SIZE
	}

	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:              util.ResourceCPUFormatter(actualRequestCPU),
			corev1.ResourceMemory:           util.ResourceMemoryFormatter(actualRequestMem),
			corev1.ResourceEphemeralStorage: resource.MustParse(k8sapi.EphemeralStorageSizeRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:              util.ResourceCPUFormatter(int(1000 * limitCPU)),
			corev1.ResourceMemory:           util.ResourceMemoryFormatter(int(limitMem)),
			corev1.ResourceEphemeralStorage: maxEphemeral,
		},
	}, nil
}
