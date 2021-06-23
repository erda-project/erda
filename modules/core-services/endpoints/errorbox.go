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
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// CreateOrUpdateErrorLog 记录或更新错误日志
func (e *Endpoints) CreateOrUpdateErrorLog(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 鉴权，创建接口只允许内部调用
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		return apierrors.ErrAddErrorLog.AccessDenied().ToResp(), nil
	}
	// 检查body是否为空
	if r.Body == nil {
		return apierrors.ErrAddErrorLog.MissingParameter("body").ToResp(), nil
	}
	// 检查body格式
	var req apistructs.ErrorLogCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrAddErrorLog.InvalidParameter(err).ToResp(), nil
	}

	// 检查事件创建请求是否合法
	if err := req.Check(); err != nil {
		return apierrors.ErrAddErrorLog.InvalidParameter(err).ToResp(), nil
	}

	// 创建信息至DB
	if err := e.errorbox.CreateOrUpdate(req); err != nil {
		return apierrors.ErrAddErrorLog.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("add error log succ")
}

// BatchCreateErrorLog 批量记录错误日志
func (e *Endpoints) BatchCreateErrorLog(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	return httpserver.OkResp("add audit succ")
}

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
		if access, err := e.permission.CheckPermission(&req); err != nil || !access {
			return apierrors.ErrGetPublisher.AccessDenied().ToResp(), nil
		}
	}

	errorLogs, err := e.errorbox.List(&listReq)
	if err != nil {
		return apierrors.ErrListErrorLog.InternalError(err).ToResp(), nil
	}

	l := len(errorLogs)
	errorLogList := make([]apistructs.ErrorLog, 0, l)
	for _, item := range errorLogs {
		errorLog := apistructs.ErrorLog{
			ID:             item.ID,
			Level:          item.Level,
			ResourceType:   item.ResourceType,
			ResourceID:     item.ResourceID,
			OccurrenceTime: item.OccurrenceTime.Format("2006-01-02 15:04:05"),
			HumanLog:       item.HumanLog,
			PrimevalLog:    item.PrimevalLog,
		}
		errorLogList = append(errorLogList, errorLog)
	}

	return httpserver.OkResp(apistructs.ErrorLogListResponseData{List: errorLogList})
}
