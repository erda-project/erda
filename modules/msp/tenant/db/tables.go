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

package db

import "time"

const (
	TableMSPTenant = "msp_tenant"
)

type MSPTenant struct {
	Id               string    `gorm:"column:id" db:"id" json:"id" form:"id"`                                                                 // Tenant id
	Type             string    `gorm:"column:type" db:"type" json:"type" form:"type"`                                                         // Tenant type（dop 、msp）
	RelatedProjectId int64     `gorm:"column:related_project_id" db:"related_project_id" json:"related_project_id" form:"related_project_id"` // Project id
	RelatedWorkspace string    `gorm:"column:related_workspace" db:"related_workspace" json:"related_workspace" form:"related_workspace"`     // Workspace（ DEV、TEST、STAGING、PROD、DEFAULT）
	CreateTime       time.Time `gorm:"column:create_time" db:"create_time" json:"create_time" form:"create_time"`                             // Create time
	UpdateTime       time.Time `gorm:"column:update_time" db:"update_time" json:"update_time" form:"update_time"`                             // Update time
	IsDeleted        bool      `gorm:"column:is_deleted" db:"is_deleted" json:"is_deleted" form:"is_deleted"`                                 // Delete or not
}

func (MSPTenant) TableName() string { return TableMSPTenant }
