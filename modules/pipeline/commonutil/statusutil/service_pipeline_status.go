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

package statusutil

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func CalculatePipelineTaskStatus(pt *spec.PipelineTask) (apistructs.PipelineStatus, error) {
	if pt.Status == apistructs.PipelineEmptyStatus {
		return apistructs.PipelineEmptyStatus, errors.New("status is empty")
	}
	if pt.Status.IsEndStatus() && pt.Extra.AllowFailure {
		return apistructs.PipelineStatusSuccess, nil // 相当于 success，并不改变数据库里原始状态
	}
	return pt.Status, nil
}

func CalculatePipelineStageStatus(ps *spec.PipelineStageWithTask) (apistructs.PipelineStatus, error) {
	tasks := ps.PipelineTasks
	totalNum := len(tasks)
	if totalNum == 0 {
		return apistructs.PipelineStatusSuccess, nil
	}

	summary := make(map[apistructs.PipelineStatus]int)
	var bornNum int
	var successNum int
	var endNum int
	var pauseNum int
	var disableNum int
	for _, task := range tasks {
		taskStatus, err := CalculatePipelineTaskStatus(task)
		if err != nil {
			return apistructs.PipelineEmptyStatus, err
		}
		summary[taskStatus]++
		if taskStatus == apistructs.PipelineStatusBorn {
			bornNum++
		}
		if taskStatus.IsSuccessStatus() {
			successNum++
		}
		if taskStatus == apistructs.PipelineStatusPaused {
			pauseNum++
		}
		if taskStatus == apistructs.PipelineStatusDisabled {
			disableNum++
		}
		if taskStatus.IsEndStatus() {
			endNum++
		}
	}

	// 重试失败节点时，可能会有 success 和 born
	// thruster 推进时需要 stage 状态为 born
	if bornNum > 0 && bornNum+successNum+disableNum == totalNum {
		return apistructs.PipelineStatusBorn, nil
	}

	// 如果全部是 终态 + 暂停
	if endNum+pauseNum+disableNum == totalNum {
		// 如果有暂停，则为暂停
		if pauseNum > 0 {
			return apistructs.PipelineStatusPaused, nil
		}
		// 禁用视为成功
		if disableNum+successNum == totalNum {
			return apistructs.PipelineStatusSuccess, nil
		}
		// 否则为失败
		return apistructs.PipelineStatusFailed, nil
	}

	// 其余状态显示为运行中
	return apistructs.PipelineStatusRunning, nil
}

// stage 递归调用 stage.CalculateStatus
func CalculateStatus(p spec.PipelineWithStage) (apistructs.PipelineStatus, error) {
	stages := p.PipelineStages
	totalNum := len(stages)
	if totalNum == 0 {
		return apistructs.PipelineStatusSuccess, nil
	}

	var stageStatusList []apistructs.PipelineStatus
	for _, stage := range stages {
		stageStatus, err := CalculatePipelineStageStatus(stage)
		if err != nil {
			return apistructs.PipelineEmptyStatus, err
		}
		stageStatusList = append(stageStatusList, stageStatus)
	}

	// 按照 stage 顺序遍历：
	// 如果有暂停，则返回暂停；
	// 如果 stage 为禁用，则查看下一个 stage；
	// 如果 stage 为终态，如果成功，则查看下一个 stage；如果失败，则快速失败
	// 返回运行中
	for si, stageStatus := range stageStatusList {
		// 如果有暂停，则为暂停
		if stageStatus == apistructs.PipelineStatusPaused {
			return apistructs.PipelineStatusPaused, nil
		}
		if stageStatus == apistructs.PipelineStatusDisabled {
			// 如果是最后一个 stage，则返回成功
			if (si + 1) == totalNum {
				return apistructs.PipelineStatusSuccess, nil
			}
			// 否则继续查看下一个 stage
			continue
		}
		if stageStatus.IsEndStatus() {
			if stageStatus.IsSuccessStatus() {
				// 如果是最后一个 stage，则返回成功
				if (si + 1) == totalNum {
					return apistructs.PipelineStatusSuccess, nil
				}
				// 否则继续查看下一个 stage
				continue
			}
			return stageStatus, nil
		}
		return apistructs.PipelineStatusRunning, nil
	}

	// 算不出来，则返回 Error
	return apistructs.PipelineStatusError, errors.Errorf("failed to calculate status of strange pipeline [%d], summary: %+v", p.ID, stageStatusList)
}
