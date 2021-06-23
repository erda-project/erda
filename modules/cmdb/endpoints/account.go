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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"

	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
)

// CreateCloudAccount 创建账号
func (e *Endpoints) CreateCloudAccount(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateCloudAccount.InvalidParameter(err).ToResp(), nil
	}

	// 从 Header 获取 OrgID
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrCreateCloudAccount.InvalidParameter("orgId is empty").ToResp(), nil
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrCreateCloudAccount.InvalidParameter("orgId is invalid").ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrCreateCloudAccount.MissingParameter("body is nil").ToResp(), nil
	}

	var accountCreateReq apistructs.CloudAccountCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&accountCreateReq); err != nil {
		return apierrors.ErrCreateCloudAccount.InvalidParameter(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: apistructs.CloudAccountResource,
		Action:   apistructs.OperateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrCreateCloudAccount.AccessDenied().ToResp(), nil
	}

	a, err := e.cloudaccount.Create(orgID, &accountCreateReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(a)
}

// UpdateCloudAccount 更新云账号
func (e *Endpoints) UpdateCloudAccount(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateProject.NotLogin().ToResp(), nil
	}

	// 检查 accountID 合法性
	accountID, err := strutil.Atoi64(vars["accountID"])
	if err != nil {
		return apierrors.ErrUpdateCloudAccount.InvalidParameter(err).ToResp(), nil
	}

	// 从 Header 获取 OrgID
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrUpdateCloudAccount.InvalidParameter("orgId is empty").ToResp(), nil
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrUpdateCloudAccount.InvalidParameter("orgId is invalid").ToResp(), nil
	}

	// 检查request body合法性
	if r.Body == nil {
		return apierrors.ErrUpdateProject.MissingParameter("body").ToResp(), nil
	}
	var accountUpdateReq apistructs.CloudAccountUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&accountUpdateReq); err != nil {
		return apierrors.ErrUpdateCloudAccount.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", accountUpdateReq)

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: apistructs.CloudAccountResource,
		Action:   apistructs.OperateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrUpdateProject.AccessDenied().ToResp(), nil
	}

	account, err := e.cloudaccount.Update(orgID, accountID, &accountUpdateReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(account)
}

// DeleteCloudAccount 删除云账号
func (e *Endpoints) DeleteCloudAccount(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDeleteCloudAccount.NotLogin().ToResp(), nil
	}

	// 检查accountID合法性
	accountID, err := strutil.Atoi64(vars["accountID"])
	if err != nil {
		return apierrors.ErrDeleteCloudAccount.InvalidParameter(err).ToResp(), nil
	}

	// 从 Header 获取 OrgID
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrDeleteCloudAccount.InvalidParameter("orgId is empty").ToResp(), nil
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrDeleteCloudAccount.InvalidParameter("orgId is invalid").ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: apistructs.CloudAccountResource,
		Action:   apistructs.OperateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrDeleteCloudAccount.AccessDenied().ToResp(), nil
	}

	// 删除云账号
	if err = e.cloudaccount.Delete(orgID, accountID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(accountID)
}

// ListCloudAccount 云账号列表/云账号查询
func (e *Endpoints) ListCloudAccount(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListCloudAccount.NotLogin().ToResp(), nil
	}

	// 从 Header 获取 OrgID
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrListCloudAccount.InvalidParameter("orgId is empty").ToResp(), nil
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrListCloudAccount.InvalidParameter("orgId is ininvalid").ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  uint64(orgID),
		Resource: apistructs.CloudAccountResource,
		Action:   apistructs.OperateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrListCloudAccount.AccessDenied().ToResp(), nil
	}

	accounts, err := e.cloudaccount.ListByOrgID(orgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(accounts)
}

// GetCloudAccount 云账号列表/云账号查询
func (e *Endpoints) GetCloudAccount(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		return apierrors.ErrCloseTicket.MissingParameter("internal client header").ToResp(), nil
	}

	// 检查 accountID 合法性
	accountID, err := strutil.Atoi64(vars["accountID"])
	if err != nil {
		return apierrors.ErrGetCloudAccount.InvalidParameter(err).ToResp(), nil
	}

	// 从 Header 获取 OrgID
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrGetCloudAccount.InvalidParameter("orgId is empty").ToResp(), nil
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetCloudAccount.InvalidParameter("orgId is invalid").ToResp(), nil
	}

	account, err := e.cloudaccount.GetByID(orgID, accountID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(account)
}
