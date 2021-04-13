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

package models

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar/pkg/gitmodule"
)

func (svc *Service) CheckPermission(repo *gitmodule.Repository, user *User, permission Permission, resourceRoleList []string) error {
	//用于pipeline的特殊用户
	if user.IsInnerUser() {
		return nil
	}
	resourceRole := ""
	if len(resourceRoleList) > 0 {
		resourceRole = strings.Join(resourceRoleList, ",")
	}
	checkPermission, err := svc.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:       user.Id,
		Scope:        "app",
		ScopeID:      uint64(repo.ApplicationId),
		Resource:     "repo",
		Action:       string(permission),
		ResourceRole: resourceRole,
	})
	if err != nil {
		return err
	}
	if !checkPermission.Access {
		return fmt.Errorf("no permission: %s for user: %s", permission, user.NickName)
	}
	return nil
}
