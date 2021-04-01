package model

import "time"

// ConfigItem 配置信息
type ConfigItem struct {
	ID           int64     `json:"id" gorm:"primary_key"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:create_time"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"column:update_time"`
	IsSync       bool      // deprecated
	Dynamic      bool      // deprecated
	Encrypt      bool      // deprecated
	DeleteRemote bool      // deprecated
	IsDeleted    string
	NamespaceID  uint64 `gorm:"index:namespace_id"`
	ItemKey      string
	ItemValue    string
	ItemComment  string
	ItemType     string // FILE, ENV
	Source       string
	Status       string // deprecated
}

// TableName 设置模型对应数据库表名称
func (ConfigItem) TableName() string {
	return "dice_config_item"
}
