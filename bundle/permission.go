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

package bundle

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// CheckPermission 鉴权
func (b *Bundle) CheckPermission(req *apistructs.PermissionCheckRequest) (*apistructs.PermissionCheckResponseData, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var permissionResp apistructs.PermissionCheckResponse
	resp, err := hc.Post(host).Path("/api/permissions/actions/check").
		JSONBody(req).
		Do().JSON(&permissionResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !permissionResp.Success {
		return nil, toAPIError(resp.StatusCode(), permissionResp.Error)
	}

	return &permissionResp.Data, nil
}

// StateCheckPermission 鉴权
func (b *Bundle) StateCheckPermission(req *apistructs.PermissionCheckRequest) (*apistructs.StatePermissionCheckResponseData, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var permissionResp apistructs.StatePermissionCheckResponse
	resp, err := hc.Post(host).Path("/api/permissions/actions/stateCheck").
		JSONBody(req).
		Do().JSON(&permissionResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !permissionResp.Success {
		return nil, toAPIError(resp.StatusCode(), permissionResp.Error)
	}

	return &permissionResp.Data, nil
}

// ScopeRoleAccess 查询给定用户是否有相应权限
func (b *Bundle) ScopeRoleAccess(userID string, req *apistructs.ScopeRoleAccessRequest) (*apistructs.ScopeRole, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var permissionResp apistructs.ScopeRoleAccessResponse
	resp, err := hc.Post(host).Path("/api/permissions/actions/access").
		Header(httputil.UserHeader, userID).
		JSONBody(req).
		Do().JSON(&permissionResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !permissionResp.Success {
		return nil, toAPIError(resp.StatusCode(), permissionResp.Error)
	}

	return &permissionResp.Data, nil
}

// ListScopeRole 获取给定用户所有角色权限
func (b *Bundle) ListScopeRole(userID, orgID string) (*apistructs.ScopeRoleList, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var permissionResp apistructs.ScopeRoleListResponse
	resp, err := hc.Get(host).Path("/api/permissions").
		Header(httputil.UserHeader, userID).
		Header(httputil.OrgHeader, orgID).
		Do().JSON(&permissionResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !permissionResp.Success {
		return nil, toAPIError(resp.StatusCode(), permissionResp.Error)
	}

	return &permissionResp.Data, nil
}
