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
	"strings"
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

// ProjectQuota is the table "ps_group_projects_quota"
// CPU quota unit is Core * 10^-3
// Mem quota uint is Byte
type ProjectQuota struct {
	ID        uint64    `gorm:"id" json:"id"`
	CreatedAt time.Time `gorm:"created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at" json:"updated_at"`

	ProjectID          uint64 `gorm:"project_id" json:"project_id"`
	ProjectName        string `gorm:"project_name" json:"project_name"`
	ProdClusterName    string `gorm:"prod_cluster_name" json:"prod_cluster_name"`
	StagingClusterName string `gorm:"staging_cluster_name" json:"staging_cluster_name"`
	TestClusterName    string `gorm:"test_cluster_name" json:"test_cluster_name"`
	DevClusterName string `gorm:"dev_cluster_name" json:"dev_cluster_name"`

	ProdCPUQuota    uint64 `gorm:"prod_cpu_quota" json:"prod_cpu_quota"`
	ProdMemQutoa    uint64 `gorm:"prod_mem_quota" json:"prod_mem_qutoa"`
	StagingCPUQuota uint64 `gorm:"staging_cpu_quota" json:"staging_cpu_quota"`
	StagingMemQuota uint64 `gorm:"staging_mem_quota" json:"staging_mem_quota"`
	TestCPUQuota    uint64 `gorm:"test_cpu_quota" json:"test_cpu_quota"`
	TestMemQuota    uint64 `gorm:"test_mem_quota" json:"test_mem_quota"`
	DevCPUQuota     uint64 `gorm:"dev_cpu_quota" json:"dev_cpu_quota"`
	DevMemQuota     uint64 `gorm:"dev_mem_quota" json:"dev_mem_quota"`

	CreatorID string `gorm:"creator_id" json:"creator_id"`
	UpdaterID string `gorm:"updater_id" json:"updater_id"`
}

// TableName returns the model's name "ps_group_projects_quota"
func (ProjectQuota) TableName() string {
	return "ps_group_projects_quota"
}

func (p ProjectQuota) GetClusterName(workspace string) string {
	switch strings.ToLower(workspace) {
	case "prod":
		return p.ProdClusterName
	case "staging":
		return p.StagingClusterName
	case "test":
		return p.TestClusterName
	case "dev":
		return p.DevClusterName
	default:
		return ""
	}
}

// GetCPUQuota returns the CPU quota on the workspace.
// The unit is Core * 10^-3
func (p ProjectQuota) GetCPUQuota(workspace string) uint64 {
	switch strings.ToLower(workspace) {
	case "prod":
		return p.ProdCPUQuota
	case "staging":
		return p.StagingCPUQuota
	case "test":
		return p.TestCPUQuota
	case "dev":
		return p.DevCPUQuota
	default:
		return 0
	}
}

// GetMemQuota returns the Mem quota on the workspace.
// The unit is Byte
func (p ProjectQuota) GetMemQuota(workspace string) uint64 {
	switch strings.ToLower(workspace) {
	case "prod":
		return p.ProdMemQutoa
	case "staging":
		return p.StagingMemQuota
	case "test":
		return p.TestMemQuota
	case "dev":
		return p.DevMemQuota
	default:
		return 0
	}
}