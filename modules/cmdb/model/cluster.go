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

// Cluster 集群数据模型
type Cluster struct {
	BaseModel
	OrgID               int64
	Name                string
	DisplayName         string
	CloudVendor         string
	Description         string
	Type                string
	Logo                string
	WildcardDomain      string
	URLs                string `gorm:"type:text"`
	Settings            string `gorm:"type:text"`
	Config              string `gorm:"type:text"` // TODO 废弃 urls & settings 字段，统一存储于 config
	SchedulerConfig     string `gorm:"column:scheduler;type:text"`
	OpsConfig           string `gorm:"column:opsconfig;type:text"`
	CloudResourceConfig string `gorm:"column:resource;type:text"`
	SysConfig           string `gorm:"column:sys;type:text"`
	ManageConfig        string `gorm:"column:manage_config;type:text"`
}

// TableName 设置模型对应数据库表名称
func (Cluster) TableName() string {
	return "co_clusters"
}
