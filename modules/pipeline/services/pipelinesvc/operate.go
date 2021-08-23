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

package pipelinesvc

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

type OperateAction string

var (
	OpDisableTask OperateAction = "DISABLE-TASK"
	OpEnableTask  OperateAction = "ENABLE-TASK"
	OpPauseTask   OperateAction = "PAUSE-TASK"
	OpUnpauseTask OperateAction = "UNPAUSE-TASK"
)

func (s *PipelineSvc) Operate(pipelineID uint64, req *apistructs.PipelineOperateRequest) error {
	p, err := s.dbClient.GetPipeline(pipelineID)
	if err != nil {
		return apierrors.ErrOperatePipeline.InternalError(err)
	}

	// var needUpdatePipelineCtx bool

	// operate task
	for _, taskOp := range req.TaskOperates {
		task, err := s.dbClient.GetPipelineTask(taskOp.TaskID)
		if err != nil {
			return err
		}

		var opAction OperateAction

		wrapError := func(err error) error {
			return errors.Wrapf(err, "failed to operate pipeline, task [%d], action [%s]", taskOp.TaskID, opAction)
		}

		// disable: pipeline 开始后无法修改
		if taskOp.Disable != nil {
			if !p.Status.CanEnableDisable() {
				return wrapError(errors.New("pipeline is already started"))
			}
			if *taskOp.Disable {
				opAction = OpDisableTask
				if !(task.Status == apistructs.PipelineStatusAnalyzed || task.Status == apistructs.PipelineStatusPaused) {
					return wrapError(errors.Errorf("invalid status [%v]", task.Status))
				}
				task.Status = apistructs.PipelineStatusDisabled
			} else {
				opAction = OpEnableTask
				task.Status = apistructs.PipelineStatusAnalyzed
			}
			// needUpdatePipelineCtx = true
		}

		// pause: task 开始执行后无法修改
		if taskOp.Pause != nil {
			if *taskOp.Pause {
				opAction = OpPauseTask
				if !task.Status.CanPause() {
					return wrapError(errors.Errorf("status [%s]", task.Status))
				}
				task.Status = apistructs.PipelineStatusPaused
			} else {
				opAction = OpUnpauseTask
				if !task.Status.CanUnpause() {
					return wrapError(errors.Errorf("status [%s]", task.Status))
				}
				// 判断当前节点所属的 stage 状态:
				// 1. 如果为 Born，则说明 stage 可以被推进器推进，则 task.status = Born 即可
				// 2. 否则，说明 stage 已经开始执行，不会被再次推进，因此 task.status = Mark
				stage, err := s.dbClient.GetPipelineStage(task.StageID)
				if err != nil {
					return err
				}
				if stage.Status == apistructs.PipelineStatusBorn {
					task.Status = apistructs.PipelineStatusBorn
				} else {
					task.Status = apistructs.PipelineStatusMark
				}
			}
			task.Extra.Pause = *taskOp.Pause
		}

		if err = s.dbClient.UpdatePipelineTask(task.ID, &task); err != nil {
			return wrapError(err)
		}
	}

	// TODO 只要在 run 里更新？
	// if needUpdatePipelineCtx {
	// 	if err := s.modifyPipelineContext(pipelineID); err != nil {
	// 		return apierrors.ErrOperatePipeline.InternalError(err)
	// 	}
	// }

	return nil
}
