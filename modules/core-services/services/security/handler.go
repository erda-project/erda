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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/types"
	"github.com/erda-project/erda/pkg/strutil"
)

type PermissionHandler interface {
	Access(ctx context.Context) bool
	PermissionListProcess(ctx context.Context) PermissionProcess
	CheckProcess(ctx context.Context) PermissionProcess
}

type InternalUserPermissionHandler struct {
	Adaptor *PermissionAdaptor
}

func (a InternalUserPermissionHandler) Access(ctx context.Context) bool {
	return internalUser(a.Adaptor.GetUserID(ctx))
}

func (a InternalUserPermissionHandler) CheckProcess(ctx context.Context) PermissionProcess {
	return PermissionProcessMiddleware{RealProcess: DefaultPermissionProcess{}, AfterCheck: func(ctx context.Context, check bool) (bool, error) {
		return true, nil
	}}
}

func (a InternalUserPermissionHandler) PermissionListProcess(ctx context.Context) PermissionProcess {
	return PermissionProcessMiddleware{RealProcess: DefaultPermissionProcess{}, AfterProcess: func(ctx context.Context, list *apistructs.PermissionList) (*apistructs.PermissionList, error) {
		return nil, fmt.Errorf("internalUserPermissionHandler not support permissionListProcess")
	}}
}

type AdminUserPermissionHandler struct {
	Adaptor *PermissionAdaptor
}

func (a AdminUserPermissionHandler) Access(ctx context.Context) bool {
	return isAdmin(a.Adaptor.Db, a.Adaptor.GetUserID(ctx))
}

func (a AdminUserPermissionHandler) CheckProcess(ctx context.Context) PermissionProcess {
	return PermissionProcessMiddleware{RealProcess: DefaultPermissionProcess{}, AfterCheck: func(ctx context.Context, check bool) (bool, error) {
		return true, nil
	}}
}

func (a AdminUserPermissionHandler) PermissionListProcess(ctx context.Context) PermissionProcess {
	if a.Adaptor.GetScopeType(ctx) == apistructs.SysScope {
		return PermissionProcessMiddleware{RealProcess: DefaultPermissionProcess{}, AfterProcess: func(ctx context.Context, list *apistructs.PermissionList) (*apistructs.PermissionList, error) {
			if list != nil {
				list.Access = true
			}
			return list, nil
		}}
	} else if a.Adaptor.GetScopeType(ctx) == apistructs.AppScope {
		return AppPermissionProcess{Adaptor: a.Adaptor}
	} else {
		return GetByUserAndScopePermissionProcess{
			Adaptor: a.Adaptor,
		}
	}
}

type SupportUserPermissionHandler struct {
	Adaptor *PermissionAdaptor
}

func (s SupportUserPermissionHandler) Access(ctx context.Context) bool {
	return s.Adaptor.GetUserID(ctx) == apistructs.SupportID
}

func (s SupportUserPermissionHandler) CheckProcess(ctx context.Context) PermissionProcess {
	return RolePermissionProcess{
		roles:   []string{types.RoleOrgSupport},
		Adaptor: s.Adaptor,
	}
}

func (s SupportUserPermissionHandler) PermissionListProcess(ctx context.Context) PermissionProcess {
	if s.Adaptor.GetScopeType(ctx) == apistructs.SysScope {
		return DefaultPermissionProcess{}
	} else {
		return RolePermissionProcess{
			roles:   []string{types.RoleOrgSupport},
			Adaptor: s.Adaptor,
		}
	}
}

type NormalUserPermissionHandler struct {
	Adaptor *PermissionAdaptor
}

func (n NormalUserPermissionHandler) Access(ctx context.Context) bool {
	return !isAdmin(n.Adaptor.Db, n.Adaptor.GetUserID(ctx)) && n.Adaptor.GetUserID(ctx) != apistructs.SupportID
}

func (n NormalUserPermissionHandler) CheckProcess(ctx context.Context) PermissionProcess {
	if n.Adaptor.GetScopeType(ctx) == apistructs.AppScope {
		return AppPermissionProcess{Adaptor: n.Adaptor}
	} else {
		return GetByUserAndScopePermissionProcess{
			Adaptor: n.Adaptor,
		}
	}
}

func (n NormalUserPermissionHandler) PermissionListProcess(ctx context.Context) PermissionProcess {
	if n.Adaptor.GetScopeType(ctx) == apistructs.AppScope {
		return AppPermissionProcess{Adaptor: n.Adaptor}
	} else {
		return GetByUserAndScopePermissionProcess{
			Adaptor: n.Adaptor,
		}
	}
}

func isAdmin(Db *dao.DBClient, userID string) bool {
	if admin, err := Db.IsSysAdmin(userID); err == nil && admin {
		return true
	} else {
		return false
	}
}

func internalUser(userID string) bool {
	if v, err := strutil.Atoi64(userID); err == nil {
		if v > 1000 && v < 5000 && userID != apistructs.SupportID {
			return true
		}
	}
	return false
}
