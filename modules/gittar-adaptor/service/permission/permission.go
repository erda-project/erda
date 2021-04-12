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

package permission

import (
	"errors"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
)

type Permission struct {
	bdl *bundle.Bundle
}

type Option func(*Permission)

// New Permission
func New(options ...Option) *Permission {
	p := &Permission{}
	for _, op := range options {
		op(p)
	}

	return p
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(p *Permission) {
		p.bdl = bdl
	}
}

func (p *Permission) check(identityInfo apistructs.IdentityInfo, req *apistructs.PermissionCheckRequest) error {
	// 内部调用，无需鉴权
	if identityInfo.InternalClient != "" {
		return nil
	}

	// 鉴权
	req.UserID = identityInfo.UserID
	respData, err := p.bdl.CheckPermission(req)
	if err != nil {
		return apierrors.ErrCheckPermission.InternalError(err)
	}
	if !respData.Access {
		return apierrors.ErrCheckPermission.AccessDenied()
	}

	return nil
}

// CheckAppAction 校验用户在应用下是否有 ${action} 权限
func (p *Permission) CheckAppAction(identityInfo apistructs.IdentityInfo, appID uint64, action string) error {
	return p.check(identityInfo, &apistructs.PermissionCheckRequest{
		Scope:    apistructs.AppScope,
		ScopeID:  appID,
		Resource: apistructs.AppResource,
		Action:   action,
	})
}

// CheckAppConfig 校验用户在应用配置管理下是否有 ${action} 权限
func (p *Permission) CheckAppConfig(identityInfo apistructs.IdentityInfo, appID uint64, action string) error {
	return p.check(identityInfo, &apistructs.PermissionCheckRequest{
		Scope:    apistructs.AppScope,
		ScopeID:  appID,
		Resource: apistructs.ConfigResource,
		Action:   action,
	})
}

// CheckBranch 校验用户在 应用对应分支 下是否有 ${action} 权限
func (p *Permission) CheckBranchAction(identityInfo apistructs.IdentityInfo, appIDStr, branch, action string) error {
	if identityInfo.IsInternalClient() {
		return nil
	}
	// appID
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrCheckPermission.InvalidParameter(err)
	}
	validBranch, err := p.bdl.GetBranchWorkspaceConfig(appID, branch)
	if err != nil {
		return apierrors.ErrCheckPermission.InternalError(err)
	}
	if validBranch.Workspace == "" {
		return apierrors.ErrCheckPermission.InternalError(errors.New("no branch rule match"))
	}

	return p.check(identityInfo, &apistructs.PermissionCheckRequest{
		Scope:    apistructs.AppScope,
		ScopeID:  appID,
		Resource: validBranch.GetPermissionResource(),
		Action:   action,
	})
}
