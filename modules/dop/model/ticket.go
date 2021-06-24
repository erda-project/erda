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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// Ticket 工单数据结构
type Ticket struct {
	dbengine.BaseModel

	Title        string
	Content      string                `json:"content" gorm:"type:text"`
	Type         apistructs.TicketType `gorm:"type:varchar(20);index:idx_type"`
	Priority     apistructs.TicketPriority
	Status       apistructs.TicketStatus
	RequestID    string `gorm:"type:varchar(60);index:idx_request_id"`
	Key          string `gorm:"type:varchar(64);index:idx_key"` // 告警使用，作为唯一 key
	OrgID        string
	Metric       string
	MetricID     string
	Count        int64 // 相同类型工单聚合
	Creator      string
	LastOperator string
	Label        string                  `json:"label" gorm:"type:text"`
	TargetType   apistructs.TicketTarget `gorm:"type:varchar(40);index:idx_target_type"`
	TargetID     string
	TriggeredAt  time.Time // 发生时间
	ClosedAt     time.Time
}

// TableName 设置模型对应数据库表名称
func (Ticket) TableName() string {
	return "ps_tickets"
}
