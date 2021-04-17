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
