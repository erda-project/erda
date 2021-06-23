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

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/modules/core-services/types"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// CheckPermission 鉴权
func (e *Endpoints) CheckPermission(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 检查请求body合法性
	if r.Body == nil {
		return apierrors.ErrCheckPermission.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.PermissionCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCheckPermission.InvalidParameter(err).ToResp(), nil
	}
	// body字段合法性校验
	if err := e.checkPermissionParam(&req); err != nil {
		return apierrors.ErrCheckPermission.InvalidParameter(err).ToResp(), nil
	}

	access, err := e.permission.CheckPermission(&req)
	if err != nil {
		return apierrors.ErrCheckPermission.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(apistructs.PermissionCheckResponseData{Access: access})
}

// StateCheckPermission 鉴权
func (e *Endpoints) StateCheckPermission(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 检查请求body合法性
	if r.Body == nil {
		return apierrors.ErrCheckPermission.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.PermissionCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCheckPermission.InvalidParameter(err).ToResp(), nil
	}

	access, roles, err := e.permission.StateCheckPermission(&req)
	if err != nil {
		return apierrors.ErrCheckPermission.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(apistructs.StatePermissionCheckResponseData{Access: access, Roles: roles})
}

func (e *Endpoints) checkPermissionParam(req *apistructs.PermissionCheckRequest) error {
	if req.UserID == "" {
		return errors.Errorf("invalid request, user id is empty")
	}

	if _, ok := types.AllScopeRoleMap[req.Scope]; !ok {
		return errors.Errorf("invalid request, scope is invalid")
	}

	if req.Resource == "" {
		return errors.Errorf("invalid request, resource is empty")
	}
	if req.Action == "" {
		return errors.Errorf("invalid request, action is empty")
	}

	return nil
}

// ListScopeRole 获取当前用户所有权限
func (e *Endpoints) ListScopeRole(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListPermission.NotLogin().ToResp(), nil
	}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrListPermission.MissingParameter("org id header").ToResp(), nil
	}
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrListPermission.InvalidParameter(err).ToResp(), nil
	}

	// 获取当前用户权限
	members, err := e.member.ListByOrgAndUser(orgID, userID.String())
	if err != nil {
		return apierrors.ErrListPermission.InternalError(err).ToResp(), nil
	}

	scopePermissionMap := make(map[string]apistructs.ScopeRole, 0)
	for _, member := range members {
		if member.ResourceKey == apistructs.RoleResourceKey {
			key := string(member.ScopeType) + strconv.FormatInt(member.ScopeID, 10)
			if v, ok := scopePermissionMap[key]; ok {
				v.Roles = append(v.Roles, member.ResourceValue)
			} else {
				scopePermissionMap[key] = apistructs.ScopeRole{
					Scope: apistructs.Scope{
						Type: member.ScopeType,
						ID:   strconv.FormatInt(member.ScopeID, 10),
					},
					Roles:  []string{member.ResourceValue},
					Access: true,
				}
			}
		}
	}

	// 数据结构转换
	permissions := make([]apistructs.ScopeRole, 0)
	for _, v := range scopePermissionMap {
		permissions = append(permissions, v)
	}
	return httpserver.OkResp(apistructs.ScopeRoleList{List: permissions})
}

// ScopeRoleAccess 根据 scope 返回对应权限列表
func (e *Endpoints) ScopeRoleAccess(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrAccessPermission.NotLogin().ToResp(), nil
	}

	// 检查请求body合法性
	if r.Body == nil {
		return apierrors.ErrAccessPermission.MissingParameter("body").ToResp(), nil
	}

	var accessReq apistructs.ScopeRoleAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&accessReq); err != nil {
		return apierrors.ErrAccessPermission.InvalidParameter("body").ToResp(), nil
	}

	// 检查Scope合法性
	if _, ok := types.AllScopeRoleMap[accessReq.Scope.Type]; !ok {
		return apierrors.ErrAccessPermission.InvalidParameter("scope type").ToResp(), nil
	}

	var scopeID int64
	if accessReq.Scope.ID != "" {
		scopeID, err = strutil.Atoi64(accessReq.Scope.ID)
		if err != nil {
			return apierrors.ErrAccessPermission.InvalidParameter("scope id").ToResp(), nil
		}
	}

	permission, err := e.getPermissionList(userID.String(), accessReq.Scope.Type, scopeID)
	if err != nil {
		return apierrors.ErrAccessPermission.InternalError(err).ToResp(), nil
	}

	// 若无权访问，返回对应 scope 的管理员信息，用于提示
	if !permission.Access {
		members, err := e.member.GetScopeManagersByScopeID(accessReq.Scope.Type, scopeID)
		if err != nil {
			return apierrors.ErrAccessPermission.InternalError(err).ToResp(), nil
		}

		// 判断scope是否已经被删除
		deleted, err := e.scopeIsDeleted(accessReq.Scope.Type, scopeID)
		if err != nil {
			return apierrors.ErrAccessPermission.InternalError(err).ToResp(), nil
		}
		if deleted {
			permission.Exist = false
		}

		for _, mem := range members {
			permission.ContactsWhenNoPermission = append(permission.ContactsWhenNoPermission, mem.UserID)
		}
	}

	return httpserver.OkResp(permission, permission.ContactsWhenNoPermission)
}

// 获取权限
func (e *Endpoints) getPermission(userID string, scopeType apistructs.ScopeType, scopeID int64) (apistructs.ScopeRole, error) {
	// 若为系统管理员 & 查询系统范围权限，则返回true；若系统管理员查询企业/项目/应用等，应返回false
	if e.member.IsAdmin(userID) && scopeType == apistructs.SysScope {
		return apistructs.ScopeRole{
			Scope: apistructs.Scope{
				Type: scopeType,
				ID:   strconv.FormatInt(scopeID, 10),
			},
			Roles:  []string{types.RoleSysManager},
			Access: true,
		}, nil
	}

	members, err := e.member.GetByUserAndScope(userID, scopeType, scopeID)
	if err != nil {
		logrus.Infof("failed to get permission, (%v)", err)
		return apistructs.ScopeRole{}, errors.Errorf("failed to access permission")
	}

	var (
		roles  []string
		access bool
	)
	if len(members) != 0 {
		for _, member := range members {
			if member.ResourceKey == apistructs.RoleResourceKey {
				roles = append(roles, member.ResourceValue)
			}
		}
		access = true
	}

	return apistructs.ScopeRole{
		Scope: apistructs.Scope{
			Type: scopeType,
			ID:   strconv.FormatInt(scopeID, 10),
		},
		Roles:  roles,
		Access: access,
	}, nil
}

// 获取权限
func (e *Endpoints) getPermissionList(userID string, scopeType apistructs.ScopeType, scopeID int64) (apistructs.PermissionList, error) {
	var permissionList = apistructs.PermissionList{
		Access:           false,
		PermissionList:   make([]apistructs.ScopeResource, 0),
		ResourceRoleList: make([]apistructs.ScopeResource, 0),
		Roles:            make([]string, 0),
		Exist:            true,
	}

	// 若为系统管理员 & 查询系统范围权限，则返回true；若系统管理员查询企业/项目/应用等，应返回false
	if e.member.IsAdmin(userID) && scopeType == apistructs.SysScope {
		permissionList.Access = true
		permissionList.Roles = []string{types.RoleSysManager}
		return permissionList, nil
	}

	var roles []string
	if userID != apistructs.SupportID {
		members, err := e.member.GetByUserAndScope(userID, scopeType, scopeID)
		if err != nil {
			logrus.Infof("failed to get permission, (%v)", err)
			return permissionList, errors.Errorf("failed to access permission")
		}

		if len(members) == 0 {
			if scopeType == apistructs.SysScope {
				return permissionList, nil
			}
			isPublic, err := e.permission.CheckPublicScope(userID, scopeType, scopeID)
			if err != nil || !isPublic {
				return permissionList, err
			}
			roles = append(roles, types.RoleGuest)
		}

		for _, member := range members {
			if member.ResourceKey == apistructs.RoleResourceKey {
				roles = append(roles, member.ResourceValue)
			}
		}
		if len(roles) == 0 {
			logrus.Warningf("nil role, scope: %s, scopeID: %d, memberInfo: %v", scopeType, scopeID, members)
			return permissionList, nil
		}
	} else {
		if scopeType == apistructs.SysScope {
			return permissionList, nil
		}
		roles = []string{types.RoleOrgSupport}
	}

	var pml = make([]model.RolePermission, 0)
	// get permissions from db
	pmlDb, resourceRoleDb := e.db.GetPermissionList(roles)

	// get permissions from permission.yml
	pmlYml, resourceRoles := conf.RolePermissions(roles)
	if len(pmlDb) > 0 {
		for _, v := range pmlDb {
			k := strutil.Concat(v.Scope, v.Resource, v.Action)
			pmlYml[k] = v
		}
	}

	// merge permission list
	for _, v := range pmlYml {
		pml = append(pml, v)
	}

	permissions := make([]apistructs.ScopeResource, 0)
	for _, v := range pml {
		permission := apistructs.ScopeResource{
			Resource:     v.Resource,
			Action:       v.Action,
			ResourceRole: v.ResourceRole,
		}
		permissions = append(permissions, permission)
	}

	resourceRoles = append(resourceRoles, resourceRoleDb...)
	resourceRoleList := make([]apistructs.ScopeResource, 0)
	for _, v := range resourceRoles {
		rr := apistructs.ScopeResource{
			Resource:     v.Resource,
			Action:       v.Action,
			ResourceRole: v.ResourceRole,
		}
		resourceRoleList = append(resourceRoleList, rr)
	}

	permissionList.Access = true
	permissionList.Roles = roles
	permissionList.PermissionList = permissions
	permissionList.ResourceRoleList = resourceRoleList
	return permissionList, nil
}

func (e *Endpoints) scopeIsDeleted(scopeType apistructs.ScopeType, scopeID int64) (bool, error) {
	switch scopeType {
	case apistructs.ProjectScope:
		_, err := e.project.Get(scopeID)
		if err != nil && err.Error() == "failed to get project: "+dao.ErrNotFoundProject.Error() {
			return true, nil
		}
		return false, err
	case apistructs.AppScope:
		_, err := e.app.Get(scopeID)
		if err != nil && err.Error() == "failed to get application: "+dao.ErrNotFoundApplication.Error() {
			return true, nil
		}
		return false, err
	}

	return false, nil
}
