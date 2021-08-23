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
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

func (s *PipelineSvc) Cancel(req *apistructs.PipelineCancelRequest) error {

	p, err := s.dbClient.GetPipeline(req.PipelineID)
	if err != nil {
		return apierrors.ErrGetPipeline.InternalError(err)
	}

	// pipeline 状态判断
	if !p.IsSnippet && !p.Status.CanCancel() {
		return errors.Errorf("invalid status [%s]", p.Status)
	}

	// 设置 cancel user
	if req.UserID != "" {
		p.Extra.CancelUser = s.tryGetUser(req.UserID)
		if err := s.dbClient.UpdatePipelineExtraExtraInfoByPipelineID(p.ID, p.Extra); err != nil {
			return err
		}
	}

	// 执行操作
	stages, err := s.dbClient.ListPipelineStageByPipelineID(p.ID)
	if err != nil {
		return err
	}
	for _, stage := range stages {
		// if !stage.Status.CanCancel() {
		// 	continue
		// }
		tasks, err := s.dbClient.ListPipelineTasksByStageID(stage.ID)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			// 嵌套任务删除流水线
			if task.IsSnippet {
				if err := s.Cancel(&apistructs.PipelineCancelRequest{
					PipelineID:   *task.SnippetPipelineID,
					IdentityInfo: req.IdentityInfo,
				}); err != nil {
					logrus.Errorf("failed to stop snippet pipeline, taskID: %d, pipelineID: %d, err: %v", task.ID, *task.SnippetPipelineID, err)
				}
				continue
			}
			if !task.Status.CanCancel() {
				continue
			}
			executor, err := actionexecutor.GetManager().Get(types.Name(task.Extra.ExecutorName))
			if err != nil {
				return err
			}
			// cancel
			if _, err = executor.Cancel(context.Background(), task); err != nil {
				return err
			}
		}
	}
	// 更新整体状态
	if err = s.dbClient.UpdateWholeStatusCancel(&p); err != nil {
		return err
	}

	// event
	events.EmitPipelineInstanceEvent(&p, req.UserID)

	return nil

}
