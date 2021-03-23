package bundle

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func toAPIError(statusCode int, err apistructs.ErrorResponse) *errorresp.APIError {
	return errorresp.New(errorresp.WithCode(statusCode, err.Code), errorresp.WithMessage(err.Msg))
}
