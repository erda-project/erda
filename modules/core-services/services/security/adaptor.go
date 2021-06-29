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

package security

import (
	"context"
	"fmt"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/dao"
)

type PermissionAdaptor struct {
	Db       *dao.DBClient
	Bdl      *bundle.Bundle
	Handlers []PermissionHandler
	Once     sync.Once
}

func (this *PermissionAdaptor) CheckPublicScope(userID string, scopeType apistructs.ScopeType, scopeID int64) (bool, error) {
	switch scopeType {
	case apistructs.OrgScope:
		org, err := this.Db.GetOrg(scopeID)
		if err != nil {
			return false, err
		}
		return org.IsPublic, nil
	case apistructs.ProjectScope:
		project, err := this.Db.GetProjectByID(scopeID)
		if err != nil || !project.IsPublic {
			return false, err
		}
		// check if in upper level
		member, err := this.Db.GetMemberByScopeAndUserID(userID, apistructs.OrgScope, project.OrgID)
		if err != nil {
			return false, err
		}
		if len(member) > 0 {
			return true, nil
		}
		// if not, check upper level isPublic
		org, err := this.Db.GetOrg(project.OrgID)
		if err != nil {
			return false, err
		}
		return org.IsPublic, nil
	case apistructs.AppScope:
		app, err := this.Db.GetApplicationByID(scopeID)
		if err != nil || !app.IsPublic {
			return false, err
		}
		member, err := this.Db.GetMemberByScopeAndUserID(userID, apistructs.ProjectScope, app.ProjectID)
		if err != nil {
			return false, err
		}
		if len(member) > 0 {
			return true, nil
		}
		project, err := this.Db.GetProjectByID(app.ProjectID)
		if err != nil || !project.IsPublic {
			return false, err
		}
		member, err = this.Db.GetMemberByScopeAndUserID(userID, apistructs.OrgScope, project.OrgID)
		if err != nil {
			return false, err
		}
		if len(member) > 0 {
			return true, nil
		}
		org, err := this.Db.GetOrg(project.OrgID)
		if err != nil {
			return false, err
		}
		return org.IsPublic, nil
	}
	return true, nil
}

func (this *PermissionAdaptor) SetCheckRequest(ctx context.Context, req apistructs.PermissionCheckRequest) context.Context {
	return context.WithValue(ctx, "checkRequest", req)
}

func (this *PermissionAdaptor) GetCheckRequest(ctx context.Context) apistructs.PermissionCheckRequest {
	return ctx.Value("checkRequest").(apistructs.PermissionCheckRequest)
}

func (this *PermissionAdaptor) SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "userID", userID)
}

func (this *PermissionAdaptor) GetUserID(ctx context.Context) string {
	return ctx.Value("userID").(string)
}

func (this *PermissionAdaptor) SetScopeType(ctx context.Context, ScopeType apistructs.ScopeType) context.Context {
	return context.WithValue(ctx, "ScopeType", ScopeType)
}

func (this *PermissionAdaptor) GetScopeType(ctx context.Context) apistructs.ScopeType {
	return ctx.Value("ScopeType").(apistructs.ScopeType)
}

func (this *PermissionAdaptor) SetScopeID(ctx context.Context, ScopeID int64) context.Context {
	return context.WithValue(ctx, "ScopeID", ScopeID)
}

func (this *PermissionAdaptor) GetScopeID(ctx context.Context) int64 {
	return ctx.Value("ScopeID").(int64)
}

func (this *PermissionAdaptor) RegisterHandler(handler PermissionHandler) {
	this.Handlers = append(this.Handlers, handler)
}

func (this *PermissionAdaptor) CheckPermission(req apistructs.PermissionCheckRequest) (bool, error) {
	var ctx = context.Background()

	ctx = this.SetScopeID(ctx, int64(req.ScopeID))
	ctx = this.SetScopeType(ctx, req.Scope)
	ctx = this.SetUserID(ctx, req.UserID)
	ctx = this.SetCheckRequest(ctx, req)

	for _, handler := range this.Handlers {
		// todo multiple handler maybe access true
		if handler.Access(ctx) {
			return handler.CheckProcess(ctx).check(ctx)
		}
	}

	return false, fmt.Errorf("not find user handler")
}

func (this *PermissionAdaptor) PermissionList(userID string, scopeType apistructs.ScopeType, scopeID int64) (apistructs.PermissionList, error) {
	var ctx = context.Background()

	ctx = this.SetScopeID(ctx, scopeID)
	ctx = this.SetScopeType(ctx, scopeType)
	ctx = this.SetUserID(ctx, userID)

	for _, handler := range this.Handlers {
		// todo multiple handler maybe access true
		if handler.Access(ctx) {
			list, err := handler.PermissionListProcess(ctx).permissionList(ctx)
			if list == nil || err != nil {
				return apistructs.PermissionList{
					Access:           false,
					PermissionList:   make([]apistructs.ScopeResource, 0),
					ResourceRoleList: make([]apistructs.ScopeResource, 0),
					Roles:            make([]string, 0),
					Exist:            true,
				}, err
			}

			return *list, err
		}
	}

	return apistructs.PermissionList{
		Access:           false,
		PermissionList:   make([]apistructs.ScopeResource, 0),
		ResourceRoleList: make([]apistructs.ScopeResource, 0),
		Roles:            make([]string, 0),
		Exist:            true,
	}, fmt.Errorf("not find user handler")
}
