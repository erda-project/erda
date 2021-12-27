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

package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// ListErrorLog 根据resource查看错误日志
func (e *Endpoints) ListErrorLog(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var listReq apistructs.ErrorLogListRequest
	if err := e.queryStringDecoder.Decode(&listReq, r.URL.Query()); err != nil {
		return apierrors.ErrListErrorLog.InvalidParameter(err).ToResp(), nil
	}

	// 查询参数检查
	if err := listReq.Check(); err != nil {
		return apierrors.ErrListErrorLog.InvalidParameter(err).ToResp(), nil
	}

	// 权限检查
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListErrorLog.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		req := apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    listReq.ScopeType,
			ScopeID:  listReq.ScopeID,
			Resource: string(listReq.ScopeType),
			Action:   apistructs.GetAction,
		}
		if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
			return apierrors.ErrListErrorLog.AccessDenied().ToResp(), nil
		}
	}

	errorLogs, err := e.taskErrorSvc.List(&listReq)
	if err != nil {
		return apierrors.ErrListErrorLog.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(apistructs.ErrorLogListResponseData{List: errorLogs})
}
