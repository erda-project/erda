// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

// Audit 审计事件
type Audit struct {
	BaseModel
	StartTime     time.Time               `gorm:"column:start_time"`
	EndTime       time.Time               `gorm:"column:end_time"`
	UserID        string                  `gorm:"column:user_id"`
	ScopeType     apistructs.ScopeType    `gorm:"column:scope_type"`
	ScopeID       uint64                  `gorm:"column:scope_id"`
	FDPProjectID  string                  `gorm:"column:fdp_project_id"`
	AppID         uint64                  `gorm:"column:app_id"`
	ProjectID     uint64                  `gorm:"column:project_id"`
	OrgID         uint64                  `gorm:"column:org_id"`
	Context       string                  `gorm:"column:context"`
	TemplateName  apistructs.TemplateName `gorm:"column:template_name"`
	AuditLevel    string                  `gorm:"column:audit_level"`
	Result        apistructs.Result       `gorm:"column:result"`
	ErrorMsg      string                  `gorm:"column:error_msg"`
	ClientIP      string                  `gorm:"column:client_ip"`
	UserAgent     string                  `gorm:"column:user_agent"`
	SoftDeletedAt uint64                  `gorm:"column:soft_deleted_at"`
}

// AuditSettings 审计事件的清理周期设置
type AuditSettings struct {
	ID     uint64
	Config OrgConfig
}

type ListAuditParam struct {
	// +required 查询事件起始时间
	StartAt string
	// +required 查询事件终止时间
	EndAt string
	// +optional 可见范围类型
	ScopeType []apistructs.ScopeType
	// +optional 企业id列表
	OrgId []uint64
	// +optional UserID列表
	UserID []string
	// +optional FDP项目ID
	FDPProjectID []string
	// +optional ScopeID列表
	ScopeID []uint64
	// +optional 应用ID列表
	AppID []uint64
	// +optional 项目ID列表
	ProjectID []uint64
	// +optional 日志模版列表
	TemplateName []apistructs.TemplateName
	// +optional 客户端IP
	ClientIp []string
	// default 1
	PageNo int
	// default 20
	PageSize int
}

// TableName 设置模型对应数据库表名称
func (Audit) TableName() string {
	return "dice_audit"
}
