package model

import "github.com/erda-project/erda/apistructs"

type MemberExtra struct {
	BaseModel
	// UeserID 关联成员的用户id
	UserID string `gorm:"column:user_id"`
	// ParentID 成员的父scope_id
	ParentID int64 `gorm:"column:parent_id"`
	// ScopeType 关联成员的scope_type
	ScopeType apistructs.ScopeType `gorm:"column:scope_type"`
	// ScopeID 关联成员的scope_id
	ScopeID int64 `gorm:"column:scope_id"`
	// ResourceKey 关联资源的key
	ResourceKey apistructs.ExtraResourceKey `gorm:"column:resource_key"`
	// ResourceValue 管理资源的值
	ResourceValue string `gorm:"column:resource_value"`
}

// TableName 表名
func (MemberExtra) TableName() string {
	return "dice_member_extra"
}
