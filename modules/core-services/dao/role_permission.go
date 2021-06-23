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

package dao

import (
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/conf"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateRolePermission 创建角色权限
func (client *DBClient) CreateRolePermission(permission *model.RolePermission) error {
	return client.Create(permission).Error
}

// GetRolePermission 获取角色权限
func (client *DBClient) GetRolePermission(roles []string, permissionInfo *apistructs.PermissionCheckRequest) (*model.RolePermission, error) {
	// 权限配置优先从数据库读取；若数据库未配置，再次尝试加载权限配置文件
	var permission model.RolePermission
	db := client.Where("scope = ?", permissionInfo.Scope).
		Where("resource = ?", permissionInfo.Resource).
		Where("action = ?", permissionInfo.Action).
		Where("role in (?)", roles)
	if permissionInfo.ResourceRole != "" {
		rrList := strings.Split(permissionInfo.ResourceRole, ",")
		for _, rr := range rrList {
			db = db.Or("resource_role LIKE ?", strutil.Concat("%", strings.TrimSpace(rr), "%"))
		}
	}

	if err := db.First(&permission).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			// 从配置文件读取权限
			pm := conf.Permissions()
			k := strutil.Concat(string(permissionInfo.Scope), permissionInfo.Resource, permissionInfo.Action)
			for _, role := range roles {
				if v, ok := pm[k]; ok {
					confRoles := strings.SplitN(v.Role, ",", -1)
					for _, confRole := range confRoles {
						if confRole == role || checkResourceRole(permissionInfo.ResourceRole, v.ResourceRole) {
							return &v, nil
						}
					}
				}
			}
			return nil, nil
		}

		return nil, err
	}
	return &permission, nil
}

// 资源的创建者或处理者的权限校验
func checkResourceRole(reqResourceRole, confResourceRole string) bool {
	if confResourceRole == "" {
		return false
	}

	for _, v := range strings.SplitN(reqResourceRole, ",", -1) {
		for _, v1 := range strings.SplitN(confResourceRole, ",", -1) {
			if v == v1 {
				return true
			}
		}
	}

	return false
}

// GetPermissionList 获取角色权限列表
func (client *DBClient) GetPermissionList(roles []string) ([]model.RolePermission, []model.RolePermission) {
	// 权限配置优先从数据库读取；若数据库未配置，再次尝试加载权限配置文件
	var (
		permissions      = make([]model.RolePermission, 0)
		roleResourceList = make([]model.RolePermission, 0)
	)
	if err := client.Where("role in (?)", roles).
		Find(&permissions).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			logrus.Warningf("failed to get permisssions, role:%v, (%+v)", roles, err)
			// 从配置文件读取权限
			return permissions, roleResourceList
		}
	}

	// 获取 resource_role不为空item
	client.Not("resource_role", "").Find(&roleResourceList)
	return permissions, roleResourceList
}
