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
	"github.com/erda-project/erda/apistructs"
)

type Line struct {
	ID uint64 `gorm:"column:id"`
}

func (client *DBClient) GetBugCountByUserID(userID uint64, projectID uint64, wontfixStateIDS []uint64, stages []string, severities []string, withReopened bool, creator uint64) (uint64, []uint64, error) {
	var lines []Line
	cli := client.Table("dice_issues").Select("id").Where("deleted = 0").Where("type = ?", apistructs.IssueTypeBug)
	if userID != 0 {
		cli = cli.Where("owner = ?", userID)
	}
	if len(wontfixStateIDS) > 0 {
		cli = cli.Where("state not in (?)", wontfixStateIDS)
	}
	if projectID != 0 {
		cli = cli.Where("project_id = ?", projectID)
	}
	if len(stages) > 0 {
		cli = cli.Where("stage in (?)", stages)
	}
	if len(severities) > 0 {
		cli = cli.Where("severity in (?)", severities)
	}
	if withReopened {
		cli = cli.Where("reopen_count > 0")
	}
	if creator != 0 {
		cli = cli.Where("creator = ?", creator)
	}
	if err := cli.Find(&lines).Error; err != nil {
		return 0, nil, err
	}

	issueIDs := lineToUint64(lines)
	return uint64(len(lines)), issueIDs, nil
}

func (client *DBClient) GetIssueNumByStatesAndUserID(ownerID, assigneeID, projectID uint64, issueType apistructs.IssueType, states, statesNotIn []uint64) (uint64, []uint64, error) {
	var lines []Line
	cli := client.Table("dice_issues").Where("deleted = 0").Select("id")
	if ownerID != 0 {
		cli = cli.Where("owner = ?", ownerID)
	}
	if assigneeID != 0 {
		cli = cli.Where("assignee = ?", assigneeID)
	}
	if projectID != 0 {
		cli = cli.Where("project_id = ?", projectID)
	}
	if len(issueType) > 0 {
		cli = cli.Where("type = ?", issueType)
	}
	if len(states) > 0 {
		cli = cli.Where("state in (?)", states)
	}
	if len(statesNotIn) > 0 {
		cli = cli.Where("state not in (?)", statesNotIn)
	}
	if err := cli.Find(&lines).Error; err != nil {
		return 0, nil, err
	}
	issueIDs := lineToUint64(lines)
	return uint64(len(lines)), issueIDs, nil
}
