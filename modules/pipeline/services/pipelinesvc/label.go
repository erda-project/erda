package pipelinesvc

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

func (s *PipelineSvc) SelectPipelineIDsByLabels(req apistructs.PipelineIDSelectByLabelRequest) ([]uint64, error) {
	pipelineIDs, err := s.dbClient.SelectPipelineIDsByLabels(req)
	if err != nil {
		return nil, apierrors.ErrSelectPipelineByLabel.InternalError(err)
	}
	return pipelineIDs, nil
}
