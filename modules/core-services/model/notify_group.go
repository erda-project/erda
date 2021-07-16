// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
