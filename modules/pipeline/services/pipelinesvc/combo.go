package pipelinesvc

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (s *PipelineSvc) AppCombos(appID uint64, req *spec.PipelineCombosReq) ([]apistructs.PipelineInvokedCombo, error) {
	combos, err := s.dbClient.ListAppInvokedCombos(appID, *req)
	if err != nil {
		return nil, apierrors.ErrListInvokedCombos.InternalError(err)
	}
	return combos, nil
}
