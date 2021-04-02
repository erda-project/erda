package model

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

// ErrorLog 错误日志模型
type ErrorLog struct {
	BaseModel
	ResourceType   apistructs.ErrorResourceType `gorm:"column:resource_type"`
	ResourceID     string                       `gorm:"column:resource_id"`
	Level          apistructs.ErrorLogLevel     `gorm:"column:level"`
	OccurrenceTime time.Time                    `gorm:"column:occurrence_time"`
	HumanLog       string                       `gorm:"column:human_log"`
	PrimevalLog    string                       `gorm:"column:primeval_log"`
	DedupID        string                       `gorm:"column:dedup_id"`
}

// TableName 设置模型对应数据库表名称
func (ErrorLog) TableName() string {
	return "dice_error_box"
}
