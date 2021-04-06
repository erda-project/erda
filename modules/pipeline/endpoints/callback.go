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
