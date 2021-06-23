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
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/pkg/database/dbengine"
)

// IssueSubscriber Issue's subscriber
type IssueSubscriber struct {
	dbengine.BaseModel

	IssueID int64
	UserID  string
}

func (IssueSubscriber) TableName() string {
	return "erda_issue_subscriber"
}

// CreateIssueSubscriber Create relationship between issue and subscriber
func (client *DBClient) CreateIssueSubscriber(issueSubscriber *IssueSubscriber) error {
	return client.Create(issueSubscriber).Error
}

// DeleteIssueSubscriber Delete issue subscriber
func (client *DBClient) DeleteIssueSubscriber(issueID int64, userID string) error {
	return client.Where("issue_id = ? and user_id = ?", issueID, userID).Delete(&IssueSubscriber{}).Error
}

// GetIssueSubscriber get subscribers of an issue
func (client *DBClient) GetIssueSubscriber(issueID int64, userID string) (*IssueSubscriber, error) {
	var issueSubscriber IssueSubscriber
	if err := client.Model(IssueSubscriber{}).Where("issue_id = ? and user_id = ?", issueID,
		userID).First(&issueSubscriber).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}

		return nil, err
	}

	return &issueSubscriber, nil
}

// GetIssueSubscribersByIssueID get subscribers by issueID
func (client *DBClient) GetIssueSubscribersByIssueID(issueID int64) ([]IssueSubscriber, error) {
	var issueSubscribers []IssueSubscriber
	if err := client.Model(IssueSubscriber{}).Where("issue_id = ?", issueID).Find(&issueSubscribers).Error; err != nil {
		return nil, err
	}

	return issueSubscribers, nil
}
