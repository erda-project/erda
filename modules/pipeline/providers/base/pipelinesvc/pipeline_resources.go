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

package pipelinesvc

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/numeral"
)

// calculatePipelineResources calculate pipeline resources according to all tasks grouped by stages.
func (s *PipelineSvc) calculatePipelineResources(allStagedTasks [][]*spec.PipelineTask) apistructs.PipelineAppliedResources {
	// merge result
	pipelineResource := apistructs.PipelineAppliedResources{
		Limits:   calculatePipelineLimitResource(allStagedTasks),
		Requests: calculatePipelineRequestResource(allStagedTasks),
	}

	return pipelineResource
}

// calculatePipelineLimitResource calculate pipeline limit resource according to all tasks grouped by stages.
//
// calculate maxResource
// 1 2 (3)
// 2 3 (5)
// 4   (4)
// => max((1+2), (2+3), (4)) = 5
func calculatePipelineLimitResource(allStagedTasks [][]*spec.PipelineTask) apistructs.PipelineAppliedResource {
	var (
		allStageCPU   []float64
		allStageMemMB []float64
	)
	for _, stageTasks := range allStagedTasks {
		var (
			stageCPU   float64 = 0
			stageMemMB float64 = 0
		)
		for _, task := range stageTasks {
			stageCPU += task.Extra.AppliedResources.Limits.CPU
			stageMemMB += task.Extra.AppliedResources.Limits.MemoryMB
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
func calculatePipelineRequestResource(allStagedTasks [][]*spec.PipelineTask) apistructs.PipelineAppliedResource {
	var allMinTaskCPUs []float64
	var allMinTaskMemMBs []float64
	for _, stageTasks := range allStagedTasks {
		for _, task := range stageTasks {
			allMinTaskCPUs = append(allMinTaskCPUs, task.Extra.AppliedResources.Requests.CPU)
			allMinTaskMemMBs = append(allMinTaskMemMBs, task.Extra.AppliedResources.Requests.MemoryMB)
		}
	}
	minStageResource := apistructs.PipelineAppliedResource{
		CPU:      numeral.MaxFloat64(allMinTaskCPUs),
		MemoryMB: numeral.MaxFloat64(allMinTaskMemMBs),
	}

	return minStageResource
}
