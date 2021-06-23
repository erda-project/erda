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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateNotice 创建平台公告
func (e *Endpoints) CreateNotice(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var createReq apistructs.NoticeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateNotice.InvalidParameter(err).ToResp(), nil
	}

	orgID, err := strconv.ParseUint(r.Header.Get(httputil.OrgHeader), 10, 64)
	if err != nil {
		return apierrors.ErrCreateNotice.InvalidParameter("orgID").ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateNotice.NotLogin().ToResp(), nil
	}

	createReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NoticeResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateNotice.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrCreateNotice.AccessDenied().ToResp(), nil
		}
	}
	noticeID, err := e.notice.Create(orgID, &createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	notice, err := e.notice.Get(noticeID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(&notice)
}

// UpdateNotice 编辑平台公告
func (e *Endpoints) UpdateNotice(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := strconv.ParseUint(r.Header.Get(httputil.OrgHeader), 10, 64)
	if err != nil {
		return apierrors.ErrUpdateNotice.InvalidParameter("orgID").ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateNotice.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateNotice.InvalidParameter(err).ToResp(), nil
	}

	var updateReq apistructs.NoticeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrUpdateNotice.InvalidParameter(err).ToResp(), nil
	}
	updateReq.ID = id
	// 用户身份

	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NoticeResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrUpdateNotice.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrUpdateNotice.AccessDenied().ToResp(), nil
		}
	}
	if err := e.notice.Update(&updateReq); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(id)
}

// PublishNotice 发布平台公告
func (e *Endpoints) PublishNotice(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := strconv.ParseUint(r.Header.Get(httputil.OrgHeader), 10, 64)
	if err != nil {
		return apierrors.ErrPublishNotice.InvalidParameter("orgID").ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPublishNotice.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrPublishNotice.InvalidParameter(err).ToResp(), nil
	}

	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NoticeResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrPublishNotice.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrPublishNotice.AccessDenied().ToResp(), nil
		}
	}
	if err := e.notice.Publish(id); err != nil {
		return errorresp.ErrResp(err)
	}

	notice, err := e.notice.Get(id)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(&notice)
}

// UnpublishNotice 下架平台公告
func (e *Endpoints) UnpublishNotice(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := strconv.ParseUint(r.Header.Get(httputil.OrgHeader), 10, 64)
	if err != nil {
		return apierrors.ErrUnpublishNotice.InvalidParameter("orgID").ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUnpublishNotice.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrUnpublishNotice.InvalidParameter(err).ToResp(), nil
	}

	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NoticeResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrUnpublishNotice.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrUnpublishNotice.AccessDenied().ToResp(), nil
		}
	}

	notice, err := e.notice.Get(id)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.notice.Unpublish(id); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(&notice)
}

// DeleteNotice 删除公告
func (e *Endpoints) DeleteNotice(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := strconv.ParseUint(r.Header.Get(httputil.OrgHeader), 10, 64)
	if err != nil {
		return apierrors.ErrDeleteNotice.InvalidParameter("orgID").ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteNotice.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteNotice.InvalidParameter(err).ToResp(), nil
	}

	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NoticeResource,
			Action:   apistructs.DeleteAction,
		})
		if err != nil {
			return apierrors.ErrDeleteNotice.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrDeleteNotice.AccessDenied().ToResp(), nil
		}
	}

	notice, err := e.notice.Get(id)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.notice.Delete(id); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(&notice)
}

// ListNotice 公告列表
func (e *Endpoints) ListNotice(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := strconv.ParseUint(r.Header.Get(httputil.OrgHeader), 10, 64)
	if err != nil {
		return apierrors.ErrListNotice.InvalidParameter("orgID").ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListNotice.NotLogin().ToResp(), nil
	}

	var listReq apistructs.NoticeListRequest
	if err := e.queryStringDecoder.Decode(&listReq, r.URL.Query()); err != nil {
		return apierrors.ErrListNotice.InvalidParameter(err).ToResp(), nil
	}
	listReq.OrgID = orgID
	listReq.IdentityInfo = identityInfo

	if !identityInfo.IsInternalClient() {
		// Authorize
		access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.NoticeResource,
			Action:   apistructs.ListAction,
		})
		if err != nil {
			return apierrors.ErrListNotice.InternalError(err).ToResp(), nil
		}
		if !access {
			return apierrors.ErrListNotice.AccessDenied().ToResp(), nil
		}
	}

	noticeResp, err := e.notice.List(&listReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// userIDs
	userIDs := make([]string, 0, len(noticeResp.List))
	for _, n := range noticeResp.List {
		userIDs = append(userIDs, n.Creator)
	}
	userIDs = strutil.DedupSlice(userIDs, true)

	return httpserver.OkResp(noticeResp, userIDs)
}
