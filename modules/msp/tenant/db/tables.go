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
	TableMSPTenant  = "msp_tenant"
	TableMSPProject = "msp_project"
)

type MSPTenant struct {
	Id               string    `gorm:"column:id" db:"id" json:"id" form:"id"`                                                                 // Tenant id
	Type             string    `gorm:"column:type" db:"type" json:"type" form:"type"`                                                         // Tenant type（dop 、msp）
	RelatedProjectId string    `gorm:"column:related_project_id" db:"related_project_id" json:"related_project_id" form:"related_project_id"` // Project id
	RelatedWorkspace string    `gorm:"column:related_workspace" db:"related_workspace" json:"related_workspace" form:"related_workspace"`     // Workspace（ DEV、TEST、STAGING、PROD、DEFAULT）
	CreatedAt        time.Time `gorm:"column:created_at" db:"created_at" json:"create_time" form:"create_time"`                               // Create time
	UpdatedAt        time.Time `gorm:"column:updated_at" db:"updated_at" json:"update_time" form:"update_time"`                               // Update time
	IsDeleted        bool      `gorm:"column:is_deleted" db:"is_deleted" json:"is_deleted" form:"is_deleted"`                                 // Delete or not
}

func (MSPTenant) TableName() string { return TableMSPTenant }

// MSPProject TODO remove msp project table
type MSPProject struct {
	Id          string    `gorm:"column:id" db:"id" json:"id" form:"id"`                                         // MSP project ID
	Name        string    `gorm:"column:name" db:"name" json:"name" form:"name"`                                 // MSP project name
	DisplayName string    `gorm:"column:display_name" db:"display_name" json:"display_name" form:"display_name"` // MSP project display name
	Type        string    `gorm:"column:type" db:"type" json:"type" form:"type"`                                 // MSP project type
	CreatedAt   time.Time `gorm:"column:created_at" db:"created_at" json:"created_at" form:"created_at"`         // Create time
	UpdatedAt   time.Time `gorm:"column:updated_at" db:"updated_at" json:"updated_at" form:"updated_at"`         // Update time
	IsDeleted   bool      `gorm:"column:is_deleted" db:"is_deleted" json:"is_deleted" form:"is_deleted"`         // Deleted or not
}

func (MSPProject) TableName() string { return TableMSPProject }
