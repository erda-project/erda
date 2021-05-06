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
