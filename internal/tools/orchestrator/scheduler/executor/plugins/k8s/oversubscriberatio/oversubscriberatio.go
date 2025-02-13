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

package oversubscriberatio

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/k8s/k8sapi"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/util"
)

const (
	DefaultRatio = 1.0
	// SUBSCRIBE_RATIO_SUFFIX the key suffix of the super ratio
	SUBSCRIBE_RATIO_SUFFIX = "_SUBSCRIBE_RATIO"
	// CPU_NUM_QUOTA cpu limit key
	CPU_NUM_QUOTA = "CPU_NUM_QUOTA"
	// CPU_CFS_PERIOD_US 100000  /sys/fs/cgroup/cpu/cpu.cfs_period_us default value
	CPU_CFS_PERIOD_US int = 100000
	// MIN_CPU_SIZE Minimum application cpu value
	MIN_CPU_SIZE = 0.1

	// MIN_MEM_SIZE Minimum application mem value
	MIN_MEM_SIZE = 10
)

type Interface interface {
	// ResourceOverCommit
	// cpu,memory field type source: apistructs/service.go.Resources
	ResourceOverCommit(workspace apistructs.DiceWorkspace, resources apistructs.Resources) (corev1.ResourceRequirements, error)
	GetOverSubscribeRatios() *OverSubscribeRatios
}

type OverSubscribeRatios struct {
	DevSubscribeRatio     *OverSubscribeRatio
	TestSubscribeRatio    *OverSubscribeRatio
	StagingSubscribeRatio *OverSubscribeRatio
	SubscribeRatio        *OverSubscribeRatio
}

type OverSubscribeRatio struct {
	CPURatio float64
	MemRatio float64
}

type provider struct {
	// Divide the CPU actually set by the upper layer by a ratio and pass it to the cluster scheduling, the default is 1
	overSubscribeRatios *OverSubscribeRatios
	// Set the cpu quota value to cpuNumQuota cpu quota, the default is 0, that is, the cpu quota is not limited
	// When the value is -1, it means that the actual number of cpus is used to set the cpu quota (quota may also be modified by other parameters, such as the number of cpus that pop up)
	cpuNumQuota float64
}

func New(options map[string]string) Interface {
	p := &provider{}
	//Get the value of the super-scoring ratio for different environments
	var (
		// Global OverSubscribeRatio
		// 1. If the OverSubscribeRatio isn't configured in the non-production workspace, we'll go with the global settings.
		//	  Otherwise, we'll stick to the workspace-specific configuration.
		// 2. Also affects in the production workspace.
		memSubscribeRatio,
		cpuSubscribeRatio,
		// OverSubscribeRatio for different non-production workspace
		devMemSubscribeRatio,
		devCpuSubscribeRatio,
		testMemSubscribeRatio,
		testCpuSubscribeRatio,
		stagingMemSubscribeRatio,
		stagingCpuSubscribeRatio float64
	)

	p.getSubscribeRatioByWorkspace(options, "PROD", "MEM", &memSubscribeRatio)
	p.getSubscribeRatioByWorkspace(options, "PROD", "CPU", &cpuSubscribeRatio)
	p.getSubscribeRatioByWorkspace(options, "DEV", "MEM", &devMemSubscribeRatio)
	p.getSubscribeRatioByWorkspace(options, "DEV", "CPU", &devCpuSubscribeRatio)
	p.getSubscribeRatioByWorkspace(options, "TEST", "MEM", &testMemSubscribeRatio)
	p.getSubscribeRatioByWorkspace(options, "TEST", "CPU", &testCpuSubscribeRatio)
	p.getSubscribeRatioByWorkspace(options, "STAGING", "MEM", &stagingMemSubscribeRatio)
	p.getSubscribeRatioByWorkspace(options, "STAGING", "CPU", &stagingCpuSubscribeRatio)

	cpuNumQuota := float64(0)
	if cpuNumQuotaValue, ok := options[CPU_NUM_QUOTA]; ok && len(cpuNumQuotaValue) > 0 {
		if num, err := strconv.ParseFloat(cpuNumQuotaValue, 64); err == nil && (num >= 0 || num == -1.0) {
			cpuNumQuota = num
			logrus.Debugf("cpuNumQuota set to %v", cpuNumQuota)
		}
	}
	return &provider{
		overSubscribeRatios: &OverSubscribeRatios{
			SubscribeRatio: &OverSubscribeRatio{
				CPURatio: cpuSubscribeRatio,
				MemRatio: memSubscribeRatio,
			},
			DevSubscribeRatio: &OverSubscribeRatio{
				CPURatio: devCpuSubscribeRatio,
				MemRatio: devMemSubscribeRatio,
			},
			TestSubscribeRatio: &OverSubscribeRatio{
				CPURatio: testCpuSubscribeRatio,
				MemRatio: testMemSubscribeRatio,
			},
			StagingSubscribeRatio: &OverSubscribeRatio{
				CPURatio: stagingCpuSubscribeRatio,
				MemRatio: stagingMemSubscribeRatio,
			},
		},
	}
}

// getSubscribeRatioByWorkspace
func (p *provider) getSubscribeRatioByWorkspace(options map[string]string, workspace string, t string, value *float64) {
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

func (p *provider) calcFineGrainedCPU(requestCPU, maxCPU, ratio float64) (float64, float64, error) {
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

	// Deprecated:
	// Reference:https://erda.cloud/terminus/dop/projects/70/apps/178/repo/tree/release/3.21/scripts/cpu_policy/policy.org
	// Processing the maximum cpu, that is, the corresponding cpu quota, the default is not limited cpu quota, that is,
	// the value corresponding to cpu.cfs_quota_us under the cgroup is -1
	//quota := k.cpuNumQuota
	//if k.cpuNumQuota == -1.0 {
	//	quota = cpupolicy.AdjustCPUSize(requestCPU)
	//}
	//
	//if quota >= requestCPU {
	//}

	return actualCPU, maxCPU, nil
}

func (p *provider) calcFineGrainedMemory(requestMem, maxMem, memSubscribeRatio float64) (float64, float64, error) {
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

// getOverSubscribeRationsByWorkspace
// Args: workspace
// Return: cpu subscribe ratio, memory subscribe ratio
func (p *provider) getOverSubscribeRationsByWorkspace(workspace apistructs.DiceWorkspace) (float64, float64) {
	subscribeRatios := map[apistructs.DiceWorkspace]*OverSubscribeRatio{
		apistructs.DevWorkspace:     p.overSubscribeRatios.DevSubscribeRatio,
		apistructs.TestWorkspace:    p.overSubscribeRatios.TestSubscribeRatio,
		apistructs.StagingWorkspace: p.overSubscribeRatios.StagingSubscribeRatio,
		apistructs.ProdWorkspace:    p.overSubscribeRatios.SubscribeRatio,
	}

	subscribeRatio, ok := subscribeRatios[workspace]
	if !ok {
		return DefaultRatio, DefaultRatio
	}

	return subscribeRatio.CPURatio, subscribeRatio.MemRatio
}

func (p *provider) ResourceOverCommit(workspace apistructs.DiceWorkspace, r apistructs.Resources) (corev1.ResourceRequirements, error) {
	// If workspace is "", use default ratio -> 1;
	// Get subscribe ratios by workspace
	cpuRatio, memRatio := p.getOverSubscribeRationsByWorkspace(workspace)

	requestCPU, limitCPU, err := p.calcFineGrainedCPU(r.Cpu, r.MaxCPU, cpuRatio)
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	requestMem, limitMem, err := p.calcFineGrainedMemory(r.Mem, r.MaxMem, memRatio)
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

func (p *provider) GetOverSubscribeRatios() *OverSubscribeRatios {
	return p.overSubscribeRatios
}
