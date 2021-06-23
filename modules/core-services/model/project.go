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
	Name           string // 项目名称
	DisplayName    string // 项目展示名称
	Desc           string // 项目描述
	Logo           string // 项目logo地址
	OrgID          int64  // 项目关联组织ID
	UserID         string `gorm:"column:creator"` // 所属用户Id
	DDHook         string `gorm:"column:dd_hook"` // 钉钉Hook
	ClusterConfig  string // 集群配置 eg: {"DEV":"terminus-y","TEST":"terminus-y","STAGING":"terminus-y","PROD":"terminus-y"}
	RollbackConfig string // 回滚配置: {"DEV": 1,"TEST": 2,"STAGING": 3,"PROD": 4}
	CpuQuota       float64
	MemQuota       float64
	Functions      string    `gorm:"column:functions"`
	ActiveTime     time.Time `gorm:"column:active_time"`
	EnableNS       bool      `gorm:"column:enable_ns"` // 是否打开项目级命名空间
	IsPublic       bool      `gorm:"column:is_public"` // 是否是公开项目
}

// TableName 设置模型对应数据库表名称
func (Project) TableName() string {
	return "ps_group_projects"
}
