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

package model

import (
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// RolePermission 角色资源操作
type RolePermission struct {
	dbengine.BaseModel
	Scope        string `gorm:"type:varchar(30);unique_index:idx_resource_action" yaml:"scope"`
	Role         string `gorm:"type:varchar(30);unique_index:idx_resource_action" yaml:"role"`
	ResourceRole string `gorm:"type:varchar(30);unique_index:idx_resource_action" yaml:"resource_role"`
	Resource     string `gorm:"type:varchar(40);unique_index:idx_resource_action" yaml:"resource"`
	Action       string `gorm:"type:varchar(30);unique_index:idx_resource_action" yaml:"action"`
	Creator      string
}

func (RolePermission) TableName() string {
	return "dice_role_permission"
}
