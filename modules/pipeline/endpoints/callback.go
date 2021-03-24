package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) pipelineCallback(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var req apistructs.PipelineCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCallback.InvalidParameter(err).ToResp(), nil
	}

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if err := e.permissionSvc.CheckInternalClient(identityInfo); err != nil {
		return errorresp.ErrResp(err)
	}

	switch req.Type {
	case string(apistructs.PipelineCallbackTypeOfAction):
		if err := e.pipelineSvc.DealPipelineCallbackOfAction(req.Data); err != nil {
			return apierrors.ErrCallback.InternalError(err).ToResp(), nil
		}
	default:
		return apierrors.ErrCallback.InvalidParameter(strutil.Concat("invalid callback type: ", req.Type)).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}
