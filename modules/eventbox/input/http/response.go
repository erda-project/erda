package http

import (
	"net/http"

	"github.com/erda-project/erda/modules/eventbox/dispatcher/errors"
	"github.com/erda-project/erda/modules/eventbox/server/types"

	"github.com/sirupsen/logrus"
)

func genResponse(dispatchErrs *errors.DispatchError) types.HTTPResponse {
	if len(dispatchErrs.BackendErrs) > 0 {
		logrus.Errorf("dispatcher backenderr: %v", dispatchErrs.BackendErrs)
		return types.HTTPResponse{Status: http.StatusBadRequest, Content: dispatchErrs.BackendErrs}
	}
	if dispatchErrs.FilterErr != nil {
		logrus.Errorf("dispatcher filterErr: %v", dispatchErrs.FilterErr)
		return types.HTTPResponse{Status: http.StatusBadRequest, Content: dispatchErrs.FilterErr.Error()}
	}
	return types.HTTPResponse{Status: http.StatusOK, Content: ""}
}
