package model

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

type MBox struct {
	BaseModel
	Title   string
	Content string
	Label   string //站内信所属模块  monitor|pipeline
	UserID  string
	Status  apistructs.MBoxStatus //read|unread
	OrgID   int64
	ReadAt  *time.Time
}

func (MBox) TableName() string {
	return "dice_mboxs"
}

func (mbox MBox) ToApiData() *apistructs.MBox {
	return &apistructs.MBox{
		ID:        mbox.ID,
		Title:     mbox.Title,
		Content:   mbox.Content,
		Label:     mbox.Label,
		Status:    mbox.Status,
		CreatedAt: mbox.CreatedAt,
		ReadAt:    mbox.ReadAt,
	}
}
