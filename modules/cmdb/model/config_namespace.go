package model

import "time"

// ConfigNamespace 配置信息
type ConfigNamespace struct {
	ID            int64     `json:"id" gorm:"primary_key"`
	CreatedAt     time.Time `json:"createdAt" gorm:"column:create_time"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"column:update_time"`
	Dynamic       bool
	IsDefault     bool
	IsDeleted     string
	Name          string
	Env           string `gorm:"index:env"`
	ProjectID     string `gorm:"index:project_id"`
	ApplicationID string `gorm:"index:application_id"`
	RuntimeID     string `gorm:"index:runtime_id"`
}

// TableName 设置模型对应数据库表名称
func (ConfigNamespace) TableName() string {
	return "dice_config_namespace"
}
