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
