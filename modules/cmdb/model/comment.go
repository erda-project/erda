package model

import "github.com/erda-project/erda/apistructs"

// Comment 工单评论模型
type Comment struct {
	BaseModel
	TicketID    int64
	CommentType apistructs.TCType
	Content     string               `gorm:"type:text"`
	IRComment   apistructs.IRComment `json:"irComment" gorm:"column:ir_comment;type:text"`
	UserID      string
}

// TableName 设置模型对应数据库表名称
func (Comment) TableName() string {
	return "ps_comments"
}
