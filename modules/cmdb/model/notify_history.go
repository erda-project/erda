package model

import (
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
)

type NotifyHistory struct {
	BaseModel
	NotifyName            string `gorm:"size:150;index:idx_notify_name"`
	NotifyItemDisplayName string `gorm:"size:150"`
	Channel               string `gorm:"size:150"`
	TargetData            string `gorm:"type:text"`
	SourceData            string `gorm:"type:text"`
	Status                string `gorm:"size:150"`
	OrgID                 int64  `gorm:"index:idx_org_id"`
	SourceType            string `gorm:"size:150"`
	SourceID              string `gorm:"size:150"`
	ErrorMsg              string `gorm:"type:text"`
	// 模块类型 cdp/workbench/monitor
	Label       string `gorm:"size:150;index:idx_module"`
	ClusterName string
}

// TableName 设置模型对应数据库表名称
func (NotifyHistory) TableName() string {
	return "dice_notify_histories"
}

func (notifyHistory *NotifyHistory) ToApiData() *apistructs.NotifyHistory {
	var (
		targets    []apistructs.NotifyTarget
		oldTargets []apistructs.OldNotifyTarget
		source     apistructs.NotifySource
	)
	if notifyHistory.TargetData != "" {
		if err := json.Unmarshal([]byte(notifyHistory.TargetData), &targets); err != nil {
			// 兼容老数据
			json.Unmarshal([]byte(notifyHistory.TargetData), &oldTargets)
			for _, v := range oldTargets {
				targets = append(targets, v.CovertToNewNotifyTarget())
			}
		}
	}

	if notifyHistory.SourceData != "" {
		json.Unmarshal([]byte(notifyHistory.SourceData), &source)
	}
	data := &apistructs.NotifyHistory{
		ID:                    notifyHistory.ID,
		NotifyName:            notifyHistory.NotifyName,
		NotifyItemDisplayName: notifyHistory.NotifyItemDisplayName,
		Channel:               notifyHistory.Channel,
		CreatedAt:             notifyHistory.CreatedAt,
		NotifyTargets:         targets,
		NotifySource:          source,
		Status:                notifyHistory.Status,
		Label:                 notifyHistory.Label,
		ErrorMsg:              notifyHistory.ErrorMsg,
	}
	return data
}
