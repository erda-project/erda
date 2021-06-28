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

package dao

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type IssueStage struct {
	dbengine.BaseModel

	OrgID     int64                `gorm:"column:org_id"`
	IssueType apistructs.IssueType `gorm:"column:issue_type"`
	Name      string               `gorm:"column:name"`
	Value     string               `gorm:"column:value"`
}

func (i IssueStage) TableName() string {
	return "dice_issue_stage"
}

func (client *DBClient) CreateIssueStage(stages []IssueStage) error {
	return client.BulkInsert(stages)
}

func (client *DBClient) DeleteIssuesStage(orgID int64, issueType apistructs.IssueType) error {
	return client.Table("dice_issue_stage").Where("org_id = ?", orgID).
		Where("issue_type = ?", issueType).Delete(IssueStage{}).Error
}

func (client *DBClient) GetIssuesStage(orgID int64, issueType apistructs.IssueType) ([]IssueStage, error) {
	var stages []IssueStage
	err := client.Table("dice_issue_stage").Where("org_id = ?", orgID).
		Where("issue_type = ?", issueType).Find(&stages).Error
	if err != nil {
		return nil, err
	}
	return stages, nil
}

// GetIssuesStageByOrgID get issuesStage by orgID
func (client *DBClient) GetIssuesStageByOrgID(orgID int64) ([]IssueStage, error) {
	var stages []IssueStage
	err := client.Table("dice_issue_stage").Where("org_id = ?", orgID).
		Find(&stages).Error
	return stages, err
}
