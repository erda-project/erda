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

package dbclient

import (
	"time"
)

// QASonar 存储sonar分析的结果，对应数据库表qa_sonar
type QASonar struct {
	ID        int64     `xorm:"pk autoincr" json:"id"`
	CreatedAt time.Time `xorm:"created" json:"createdAt"`
	UpdatedAt time.Time `xorm:"updated" json:"updatedAt"`

	ApplicationID    int64  `xorm:"app_id" json:"applicationId"`
	ProjectID        int64  `xorm:"project_id" json:"projectId" validate:"required"`
	BuildID          int64  `xorm:"build_id" json:"buildId"`
	LogID            string `xorm:"log_id" json:"logId"`
	ApplicationName  string `xorm:"app_name" json:"applicationName"`
	OperatorID       string `xorm:"operator_id" json:"operatorId" validate:"required"`
	CommitID         string `xorm:"commit_id" json:"commitId"`
	Branch           string `xorm:"branch" json:"branch" validate:"required"`
	GitRepo          string `xorm:"git_repo" json:"gitRepo" validate:"required"`
	Key              string `xorm:"not null VARCHAR(255)" json:"key,omitempty"`
	Bugs             string `xorm:"longtext" json:"bugs,omitempty"`
	CodeSmells       string `xorm:"longtext" json:"code_smells,omitempty"`
	Vulnerabilities  string `xorm:"longtext" json:"vulnerabilities,omitempty"`
	Coverage         string `xorm:"longtext" json:"coverage,omitempty"`
	Duplications     string `xorm:"longtext" json:"duplications,omitempty"`
	IssuesStatistics string `xorm:"text" json:"issues_statistics,omitempty"`
}

// TableName QASonar对应的数据库表qa_sonar
func (QASonar) TableName() string {
	return "qa_sonar"
}
