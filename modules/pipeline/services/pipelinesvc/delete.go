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
