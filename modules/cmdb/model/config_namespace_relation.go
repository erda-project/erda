package model

import "time"

// ConfigNamespaceRelation 配置信息
type ConfigNamespaceRelation struct {
	ID               int64     `json:"id" gorm:"primary_key"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:create_time"`
	UpdatedAt        time.Time `json:"updatedAt" gorm:"column:update_time"`
	IsDeleted        string
	Namespace        string `gorm:"index:namespace"`
	DefaultNamespace string `gorm:"index:default_namespace"`
}

// TableName 设置模型对应数据库表名称
func (ConfigNamespaceRelation) TableName() string {
	return "dice_config_namespace_relation"
}
