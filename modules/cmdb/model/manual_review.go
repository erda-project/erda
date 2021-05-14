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

type ManualReview struct {
	BaseModel
	BuildId         int    `gorm:"build_id"`
	ProjectId       int    `gorm:"project_id"`
	ApplicationId   int    `gorm:"application_id"`
	ApplicationName string `gorm:"application_name"`
	SponsorId       string `gorm:"sponsor_id"`
	CommitID        string `gorm:"commit_id"`
	OrgId           int64  `gorm:"org_id"`
	TaskId          int    `gorm:"task_id"`
	ProjectName     string `gorm:"project_name"`
	BranchName      string `gorm:"branch_name"`
	ApprovalStatus  string `gorm:"approval_status"`
	CommitMessage   string `gorm:"commit_message"`
	ApprovalReason  string `gorm:"approval_reason"`
}

type ReviewUser struct {
	BaseModel
	Operator string `gorm:"operator"`
	OrgId    int64  `gorm:"org_id"`
	TaskId   int64  `gorm:"task_id"`
}
