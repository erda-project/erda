package model

import "time"

// Approve 审批信息模型
type Approve struct {
	BaseModel
	OrgID        uint64
	TargetID     uint64
	EntityID     uint64
	TargetName   string
	Extra        string
	Title        string
	Priority     string
	Status       string
	Submitter    string
	Approver     string
	Type         string // IOS发布证书/Android证书/消息推送证书
	Desc         string
	ApprovalTime *time.Time
}

// TableName 设置模型对应数据库表名称
func (Approve) TableName() string {
	return "dice_approves"
}
