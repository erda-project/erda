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
)

// Audit 审计事件
type Audit struct {
	BaseModel
	StartTime    time.Time               `gorm:"column:start_time"`
	EndTime      time.Time               `gorm:"column:end_time"`
	UserID       string                  `gorm:"column:user_id"`
	ScopeType    apistructs.ScopeType    `gorm:"column:scope_type"`
	ScopeID      uint64                  `gorm:"column:scope_id"`
	FDPProjectID string                  `gorm:"column:fdp_project_id"`
	AppID        uint64                  `gorm:"column:app_id"`
	ProjectID    uint64                  `gorm:"column:project_id"`
	OrgID        uint64                  `gorm:"column:org_id"`
	Context      string                  `gorm:"column:context"`
	TemplateName apistructs.TemplateName `gorm:"column:template_name"`
	AuditLevel   string                  `gorm:"column:audit_level"`
	Result       apistructs.Result       `gorm:"column:result"`
	ErrorMsg     string                  `gorm:"column:error_msg"`
	ClientIP     string                  `gorm:"column:client_ip"`
	UserAgent    string                  `gorm:"column:user_agent"`
	Deleted      int                     `gorm:"column:deleted"`
}

// AuditSettings 审计事件的清理周期设置
type AuditSettings struct {
	ID     uint64
	Config OrgConfig
}

// TableName 设置模型对应数据库表名称
func (Audit) TableName() string {
	return "dice_audit"
}
