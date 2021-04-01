package model

type NotifySource struct {
	BaseModel
	Name       string `gorm:"size:150"`
	NotifyID   int64  `gorm:"index:notify_id"`
	SourceType string `gorm:"index:source_type"`
	SourceID   string `gorm:"index:source_id"`
	OrgID      int64
}

// TableName 设置模型对应数据库表名称
func (NotifySource) TableName() string {
	return "dice_notify_sources"
}
