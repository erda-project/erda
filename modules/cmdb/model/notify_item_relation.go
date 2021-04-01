package model

type NotifyItemRelation struct {
	BaseModel
	NotifyID     int64 `gorm:"index:notify_id"`
	NotifyItemID int64 `gorm:"index:notify_item_id"`
}

// TableName 设置模型对应数据库表名称
func (NotifyItemRelation) TableName() string {
	return "dice_notify_item_relation"
}
