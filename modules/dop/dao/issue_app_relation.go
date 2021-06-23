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

import "github.com/erda-project/erda/pkg/database/dbengine"

// IssueAppRelation 事件应用关联
type IssueAppRelation struct {
	dbengine.BaseModel

	IssueID   int64
	CommentID int64
	AppID     int64
	MRID      int64
}

// TableName 表名
func (IssueAppRelation) TableName() string {
	return "dice_issue_app_relations"
}

// CreateIssueAppRelation 创建事件应用关联关系
func (client *DBClient) CreateIssueAppRelation(issueAppRel *IssueAppRelation) error {
	return client.Create(issueAppRel).Error
}

// DeleteIssueAppRelationsByComment 根据 commentID 删除关联关系
func (client *DBClient) DeleteIssueAppRelationsByComment(commentID int64) error {
	return client.Where("comment_id = ?", commentID).Delete(&IssueAppRelation{}).Error
}

// DeleteIssueAppRelationsByApp 根据 appID 删除关联关系
func (client *DBClient) DeleteIssueAppRelationsByApp(appID int64) error {
	return client.Where("app_id = ?", appID).Delete(&IssueAppRelation{}).Error
}
