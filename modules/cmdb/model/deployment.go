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

// Deployments 部署的服务信息
type Deployments struct {
	ID              int64     `json:"id" gorm:"primary_key"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	OrgID           uint64    `gorm:"index:org_id"`
	ProjectID       uint64
	ApplicationID   uint64
	PipelineID      uint64
	TaskID          uint64
	QueueTimeSec    int64 // 排队耗时
	CostTimeSec     int64 // 任务耗时
	ProjectName     string
	ApplicationName string
	TaskName        string
	Status          string
	Env             string
	ClusterName     string
	UserID          string
	RuntimeID       string
	ReleaseID       string
	Extra           ExtraDeployment `json:"extra"`
}

// 额外字段预留
type ExtraDeployment struct{}

// TableName 设置模型对应数据库表名称
func (Deployments) TableName() string {
	return "cm_deployments"
}
