package model

// Activity 活动模型
type Activity struct {
	BaseModel
	OrgID         int64
	ProjectID     int64
	ApplicationID int64
	BuildID       int64
	RuntimeID     int64
	UserID        string `gorm:"column:operator"`
	Type          string // 活动类型
	Action        string
	Desc          string // 活动描述
	Context       string `json:"context" gorm:"type:text"`
}

// TableName 设置模型对应数据库表名称
func (Activity) TableName() string {
	return "ps_activities"
}
