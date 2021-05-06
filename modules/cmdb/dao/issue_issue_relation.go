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

import "github.com/jinzhu/gorm"

// IssueRelation 事件事件关联
type IssueRelation struct {
	BaseModel
	IssueID      uint64
	RelatedIssue uint64
	Comment      string
}

// TableName 表名
func (IssueRelation) TableName() string {
	return "dice_issue_relation"
}

// CreateIssueRelations 创建事件与事件的关联关系
func (client *DBClient) CreateIssueRelations(issueRelation *IssueRelation) error {
	return client.Create(issueRelation).Error
}

// GetRelatingIssues 获取该事件关联了哪些事件
func (client *DBClient) GetRelatingIssues(issueID uint64) ([]uint64, error) {
	var (
		issueIDs       []uint64
		issueRelations []IssueRelation
	)

	if err := client.Table("dice_issue_relation").Where("issue_id = ?", issueID).Find(&issueRelations).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, v := range issueRelations {
		issueIDs = append(issueIDs, v.RelatedIssue)
	}

	return issueIDs, nil
}

// GetRelatedIssues 获取该事件被哪些事件关联了
func (client *DBClient) GetRelatedIssues(issueID uint64) ([]uint64, error) {
	var (
		issueIDs       []uint64
		issueRelations []IssueRelation
	)

	if err := client.Table("dice_issue_relation").Where("related_issue = ?", issueID).Find(&issueRelations).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, v := range issueRelations {
		issueIDs = append(issueIDs, v.IssueID)
	}

	return issueIDs, nil
}

// DeleteIssueRelation 删除两条issue之间的关联关系
func (client *DBClient) DeleteIssueRelation(issueID, relatedIssueID uint64) error {
	if err := client.Table("dice_issue_relation").Where("issue_id = ?", issueID).
		Where("related_issue = ?", relatedIssueID).Delete(IssueRelation{}).Error; err != nil {
		return err
	}

	return nil
}

// ClearIssueRelation 清理所有的issue关联关系
func (client *DBClient) CleanIssueRelation(issueID uint64) error {
	if err := client.Table("dice_issue_relation").Where("issue_id = ? or related_issue = ?", issueID, issueID).
		Delete(IssueRelation{}).Error; err != nil {
		return err
	}

	return nil
}

func (client *DBClient) GetIssueRelationsByIDs(issueIDs []uint64) ([]IssueRelation, error) {
	var issueRelations []IssueRelation
	if err := client.Debug().Table("dice_issue_relation").Where("issue_id in (?)", issueIDs).Find(&issueRelations).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return issueRelations, nil
}
