package model

// Publisher 资源模型
type Publisher struct {
	BaseModel
	Name          string // Publisher名称
	PublisherType string // Publisher类型
	PublisherKey  string // PublisherKey，可以作为唯一标示，主要用于监控
	Desc          string // Publisher描述
	Logo          string // Publisher logo地址
	OrgID         int64  // Publisher关联组织ID
	UserID        string `gorm:"column:creator"` // 所属用户Id
}

// TableName 设置模型对应数据库表名称
func (Publisher) TableName() string {
	return "dice_publishers"
}
