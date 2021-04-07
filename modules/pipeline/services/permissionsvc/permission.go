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

package permissionsvc

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/gitflowutil"
)

type PermissionSvc struct {
	bdl *bundle.Bundle
}

func New(bdl *bundle.Bundle) *PermissionSvc {
	s := PermissionSvc{}
	s.bdl = bdl
	return &s
}

func (s *PermissionSvc) Check(identityInfo apistructs.IdentityInfo, req *apistructs.PermissionCheckRequest) error {
	// 内部调用，无需鉴权
	if identityInfo.InternalClient != "" {
		return nil
	}

	// 鉴权
	req.UserID = identityInfo.UserID
	respData, err := s.bdl.CheckPermission(req)
	if err != nil {
		return apierrors.ErrCheckPermission.InternalError(err)
	}
	if !respData.Access {
		return apierrors.ErrCheckPermission.AccessDenied()
	}

	return nil
}

func (s *PermissionSvc) CheckInternalClient(identityInfo apistructs.IdentityInfo) error {
	return s.Check(identityInfo, &apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,    // just check internal user id
		Scope:    apistructs.OrgScope,    // any valid scope is ok
		Resource: apistructs.OrgResource, // any valid resource is ok
		Action:   apistructs.GetAction,   // any valid action is ok
	})
}

// CheckApp 校验用户在 应用 下是否有 ${action} 权限
func (s *PermissionSvc) CheckApp(identityInfo apistructs.IdentityInfo, appID uint64, action string) error {
	return s.Check(identityInfo, &apistructs.PermissionCheckRequest{
		Scope:    apistructs.AppScope,
		ScopeID:  appID,
		Resource: apistructs.AppResource,
		Action:   action,
	})
}

// CheckBranch 校验用户在 应用对应分支 下是否有 ${action} 权限
func (s *PermissionSvc) CheckBranch(identityInfo apistructs.IdentityInfo, appIDStr, branch, action string) error {
	if identityInfo.IsInternalClient() {
		return nil
	}
	// 处理分支，获取分支前缀
	branchPrefix, err := gitflowutil.GetReferencePrefix(branch)
	if err != nil {
		return apierrors.ErrCheckPermission.InvalidParameter(err)
	}
	// appID
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrCheckPermission.InvalidParameter(err)
	}
	return s.Check(identityInfo, &apistructs.PermissionCheckRequest{
		Scope:    apistructs.AppScope,
		ScopeID:  appID,
		Resource: branchPrefix,
		Action:   action,
	})
}
