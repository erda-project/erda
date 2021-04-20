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
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

func (s *PipelineSvc) Delete(pipelineID uint64) error {

	// 获取 pipeline
	p, err := s.Get(pipelineID)
	if err != nil {
		return apierrors.ErrGetPipeline.InvalidParameter(err)
	}
	// 校验当前流水线是否可被删除
	can, reason := canDelete(*p)
	if !can {
		return apierrors.ErrDeletePipeline.InvalidState(reason)
	}

	// pipelines
	if err := s.dbClient.DeletePipeline(pipelineID); err != nil {
		return apierrors.ErrDeletePipeline.InternalError(err)
	}

	// related pipeline stages
	if err := s.dbClient.DeletePipelineStagesByPipelineID(pipelineID); err != nil {
		return apierrors.ErrDeletePipelineStage.InternalError(err)
	}

	// related pipeline tasks
	if err := s.dbClient.DeletePipelineTasksByPipelineID(pipelineID); err != nil {
		return apierrors.ErrDeletePipelineTask.InternalError(err)
	}

	// related pipeline labels
	if err := s.dbClient.DeletePipelineLabelsByPipelineID(pipelineID); err != nil {
		return apierrors.ErrDeletePipelineLabel.InternalError(err)
	}

	return nil
}
