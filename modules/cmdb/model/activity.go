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
