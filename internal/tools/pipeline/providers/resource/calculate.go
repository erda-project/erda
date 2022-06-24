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

package resource

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/numeral"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

// about pipeline task resource env
const (
	EnvPipelineLimitedCPU    = "PIPELINE_LIMITED_CPU"
	EnvPipelineRequestedCPU  = "PIPELINE_REQUESTED_CPU"
	EnvPipelineLimitedMem    = "PIPELINE_LIMITED_MEM"
	EnvPipelineLimitedDisk   = "PIPELINE_LIMITED_DISK"
	EnvPipelineRequestedMem  = "PIPELINE_REQUESTED_MEM"
	EnvPipelineRequestedDisk = "PIPELINE_REQUESTED_DISK"
)

// calculatePipelineLimitResource calculate pipeline limit resource according to all tasks grouped by stages.
//
// calculate maxResource
// 1 2 (3)
// 2 3 (5)
// 4   (4)
// => max((1+2), (2+3), (4)) = 5
func calculatePipelineLimitResource(resources [][]*apistructs.PipelineAppliedResources) apistructs.PipelineAppliedResource {
	var (
		allStageCPU   []float64
		allStageMemMB []float64
	)

	for _, stagesResources := range resources {
		var (
			stageCPU   float64 = 0
			stageMemMB float64 = 0
		)
		for _, taskResources := range stagesResources {
			stageCPU += taskResources.Limits.CPU
			stageMemMB += taskResources.Limits.MemoryMB
		}
		allStageCPU = append(allStageCPU, stageCPU)
		allStageMemMB = append(allStageMemMB, stageMemMB)
	}
	maxStageResource := apistructs.PipelineAppliedResource{
		CPU:      numeral.MaxFloat64(allStageCPU),
		MemoryMB: numeral.MaxFloat64(allStageMemMB),
	}

	return maxStageResource
}

// calculatePipelineRequestResource calculate pipeline request resource according to all tasks grouped by stages.
//
// calculate minResource
// 1 2 (2)
// 2 3 (3)
// 4   (4)
// => max(1,2,2,3,4) = 4
func calculatePipelineRequestResource(resources [][]*apistructs.PipelineAppliedResources) apistructs.PipelineAppliedResource {
	var allMinTaskCPUs []float64
	var allMinTaskMemMBs []float64
	for _, stagesResources := range resources {
		for _, taskResources := range stagesResources {
			allMinTaskCPUs = append(allMinTaskCPUs, taskResources.Requests.CPU)
			allMinTaskMemMBs = append(allMinTaskMemMBs, taskResources.Requests.MemoryMB)
		}
	}
	minStageResource := apistructs.PipelineAppliedResource{
		CPU:      numeral.MaxFloat64(allMinTaskCPUs),
		MemoryMB: numeral.MaxFloat64(allMinTaskMemMBs),
	}

	return minStageResource
}

// calculateOversoldTaskLimitResource cpu multiply the default oversold rate. if larger than max cpu default,use default max cpu
// TODO memory oversold
func calculateOversoldTaskLimitResource(limits apistructs.PipelineAppliedResource, overSoldRes apistructs.PipelineOverSoldResource) apistructs.PipelineAppliedResource {
	maxCPU := limits.CPU
	maxMemoryMB := limits.MemoryMB
	// Cpu is usually be wasted, if action and action define cpu is lower than default, use default cpu
	maxCPU = maxCPU * float64(overSoldRes.CPURate)
	if maxCPU > overSoldRes.MaxCPU {
		maxCPU = overSoldRes.MaxCPU
	}
	return apistructs.PipelineAppliedResource{
		CPU:      maxCPU,
		MemoryMB: maxMemoryMB,
	}
}

func calculateNormalTaskLimitResource(action *pipelineyml.Action, actionDefine *diceyml.Job, defaultRes apistructs.PipelineAppliedResource) apistructs.PipelineAppliedResource {
	// Calculate if actionDefine not empty
	var maxCPU, maxMemoryMB float64
	if actionDefine != nil {
		maxCPU = numeral.MaxFloat64([]float64{
			actionDefine.Resources.MaxCPU, actionDefine.Resources.CPU,
			action.Resources.MaxCPU, action.Resources.CPU,
		})
		maxMemoryMB = numeral.MaxFloat64([]float64{
			float64(actionDefine.Resources.MaxMem), float64(actionDefine.Resources.Mem),
			float64(action.Resources.Mem),
		})
	}

	// use default if is empty
	if maxCPU == 0 {
		maxCPU = defaultRes.CPU
	}
	if maxMemoryMB == 0 {
		maxMemoryMB = defaultRes.MemoryMB
	}

	return apistructs.PipelineAppliedResource{
		CPU:      maxCPU,
		MemoryMB: maxMemoryMB,
	}
}

func calculateNormalTaskRequestResource(action *pipelineyml.Action, actionDefine *diceyml.Job, defaultRes apistructs.PipelineAppliedResource) apistructs.PipelineAppliedResource {
	// calculate if requestCPU not empty
	var requestCPU, requestMemoryMB float64
	if actionDefine != nil {
		requestCPU = numeral.MinFloat64([]float64{actionDefine.Resources.MaxCPU, actionDefine.Resources.CPU}, true)
		requestMemoryMB = numeral.MinFloat64([]float64{float64(actionDefine.Resources.MaxMem), float64(actionDefine.Resources.Mem)}, true)
	}

	// user explicit declaration has the highest priority, overwrite value from actionDefine
	if c := numeral.MinFloat64([]float64{action.Resources.MaxCPU, action.Resources.CPU}, true); c > 0 {
		requestCPU = c
	}
	if m := action.Resources.Mem; m > 0 {
		requestMemoryMB = float64(m)
	}

	// use default if is empty
	if requestCPU == 0 {
		requestCPU = defaultRes.CPU
	}
	if requestMemoryMB == 0 {
		requestMemoryMB = defaultRes.MemoryMB
	}

	return apistructs.PipelineAppliedResource{
		CPU:      requestCPU,
		MemoryMB: requestMemoryMB,
	}
}
