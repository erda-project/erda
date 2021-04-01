package model

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

type NotifyGroup struct {
	BaseModel
	Name        string `gorm:"size:150"`
	ScopeType   string `gorm:"size:150;index:idx_scope_type"`
	ScopeID     string `gorm:"size:150;index:idx_scope_id"`
	OrgID       int64  `gorm:"index:idx_org_id"`
	TargetData  string `gorm:"type:text"`
	Label       string `gorm:"size:200"`
	ClusterName string
	AutoCreate  bool
	Creator     string `gorm:"size:150"`
}

func (NotifyGroup) TableName() string {
	return "dice_notify_groups"
}

func (notifyGroup *NotifyGroup) ToApiData() *apistructs.NotifyGroup {
	var targets []apistructs.NotifyTarget
	if notifyGroup.TargetData != "" {
		err := json.Unmarshal([]byte(notifyGroup.TargetData), &targets)
		if err != nil {
			// 老数据兼容
			targets = targets[:0]
			var oldTarget []apistructs.OldNotifyTarget
			err = json.Unmarshal([]byte(notifyGroup.TargetData), &oldTarget)
			if err != nil {
				logrus.Errorf("compatible old notify target error: %v", err)
			}
			for _, ot := range oldTarget {
				targets = append(targets, ot.CovertToNewNotifyTarget())
			}
		}
	}
	data := &apistructs.NotifyGroup{
		ID:        notifyGroup.ID,
		Name:      notifyGroup.Name,
		ScopeType: notifyGroup.ScopeType,
		ScopeID:   notifyGroup.ScopeID,
		Targets:   targets,
		CreatedAt: notifyGroup.CreatedAt,
		Creator:   notifyGroup.Creator,
	}
	return data
}
