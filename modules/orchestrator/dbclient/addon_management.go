package dbclient

import (
	"time"
)

// relay on the ops modules（ops/dbclient/addon_management.go）
// remove the dependency when ops complete the migration

// addon management
type AddonManagement struct {
	ID          uint64 `gorm:"primary_key"`
	AddonID     string `gorm:"type:varchar(64)"` // 主键
	Name        string `gorm:"type:varchar(64)"`
	ProjectID   string
	OrgID       string
	AddonConfig string `gorm:"type:text"`
	CPU         float64
	Mem         uint64
	Nodes       int
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

func (AddonManagement) TableName() string {
	return "tb_addon_management"
}
