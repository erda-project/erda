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
)

// Project 项目资源模型
type Project struct {
	BaseModel
	Name           string // Project name
	DisplayName    string // Project display name
	Desc           string // Project description
	Logo           string // Project logo address
	OrgID          int64  // Project related organization ID
	UserID         string `gorm:"column:creator"` // 所属用户Id
	DDHook         string `gorm:"column:dd_hook"` // 钉钉Hook
	ClusterConfig  string // Cluster configuration eg: {"DEV":"terminus-y","TEST":"terminus-y","STAGING":"terminus-y","PROD":"terminus-y"}
	RollbackConfig string // Rollback configuration: {"DEV": 1,"TEST": 2,"STAGING": 3,"PROD": 4}
	CpuQuota       float64
	MemQuota       float64
	Functions      string    `gorm:"column:functions"`
	ActiveTime     time.Time `gorm:"column:active_time"`
	EnableNS       bool      `gorm:"column:enable_ns"` // Whether to open the project-level namespace
	IsPublic       bool      `gorm:"column:is_public"` // Is it a public project
	Type           string    `gorm:"column:type"`      // project type
}

// TableName 设置模型对应数据库表名称
func (Project) TableName() string {
	return "ps_group_projects"
}
