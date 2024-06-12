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

package dbclient

import "time"

const (
	TableMSPTenant = "msp_tenant"
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

func (db *DBClient) QueryTenantByProjectIDAndWorkspace(projectId, workSpace string) (*MSPTenant, error) {
	tenant := MSPTenant{}
	err := db.
		Model(&tenant).
		Where("`related_project_id` = ?", projectId).
		Where("`related_workspace` = ?", workSpace).
		Find(&tenant).Error
	return &tenant, err
}
