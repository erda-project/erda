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
