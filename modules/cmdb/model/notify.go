package model

type Notify struct {
	BaseModel
	Name          string `gorm:"size:150"`
	ScopeType     string `gorm:"size:150;index:idx_scope_type"`
	ScopeID       string `gorm:"size:150;index:idx_scope_id"`
	Label         string `gorm:"size:150"`
	ClusterName   string
	Channels      string `gorm:"type:text"`
	NotifyGroupID int64  `gorm:"index:notify_group_id"`
	OrgID         int64  `gorm:"index:idx_org_id"`
	Creator       string
	Enabled       bool
	Data          string `gorm:"type:text"`
}

// TableName 设置模型对应数据库表名称
func (Notify) TableName() string {
	return "dice_notifies"
}
