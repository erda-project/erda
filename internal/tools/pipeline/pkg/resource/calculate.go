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

// CalculatePipelineLimitResource calculate pipeline limit resource according to all tasks grouped by stages.
//
// calculate maxResource
// 1 2 (3)
// 2 3 (5)
// 4   (4)
// => max((1+2), (2+3), (4)) = 5
func CalculatePipelineLimitResource(resources [][]*apistructs.PipelineAppliedResources) apistructs.PipelineAppliedResource {
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

// CalculatePipelineRequestResource calculate pipeline request resource according to all tasks grouped by stages.
//
// calculate minResource
// 1 2 (2)
// 2 3 (3)
// 4   (4)
// => max(1,2,2,3,4) = 4
func CalculatePipelineRequestResource(resources [][]*apistructs.PipelineAppliedResources) apistructs.PipelineAppliedResource {
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
