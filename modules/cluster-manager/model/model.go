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

// BaseModel common info for all models
type BaseModel struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Cluster cluster model
type Cluster struct {
	BaseModel
	OrgID           int64
	Name            string
	DisplayName     string
	CloudVendor     string
	Description     string
	Type            string
	Logo            string
	WildcardDomain  string
	SchedulerConfig string `gorm:"column:scheduler;type:text"`
	OpsConfig       string `gorm:"column:opsconfig;type:text"`
	ManageConfig    string `gorm:"column:manage_config;type:text"`
	SysConfig       string `gorm:"column:sys;type:text"`
	// Deprecated
	URLs string `gorm:"type:text"`
}

// TableName cluster table name
func (Cluster) TableName() string {
	return "co_clusters"
}
