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

package dao

import (
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type IssueStateCirculation struct {
	dbengine.BaseModel

	ProjectID uint64
	IssueID   uint64
	StateFrom uint64
	StateTo   uint64
	Creator   string
}

func (IssueStateCirculation) TableName() string {
	return "dice_issue_state_circulation"
}

func (client *DBClient) CreateIssueStateCirculation(StatesCircus *IssueStateCirculation) error {
	return client.Create(StatesCircus).Error
}

func (client *DBClient) DeleteIssuesStateCirculation(issueID uint64) error {
	return client.Model(&IssueStateCirculation{}).Where("issue_id = ?", issueID).Delete(IssueStateCirculation{}).Error
}

func (client *DBClient) ListStatesCircusByProjectID(projectID uint64) ([]IssueStateCirculation, error) {
	var statesCircus []IssueStateCirculation
	db := client.Model(&IssueStateCirculation{}).Where("project_id = ?", projectID)
	if err := db.Find(&statesCircus).Error; err != nil {
		return nil, err
	}
	return statesCircus, nil
}
