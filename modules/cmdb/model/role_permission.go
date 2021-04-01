package model

import (
	"github.com/erda-project/erda/pkg/dbengine"
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
