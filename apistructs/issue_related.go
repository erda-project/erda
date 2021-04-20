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

package apistructs

import "github.com/pkg/errors"

type IssueRelationCreateRequest struct {
	IssueID      uint64 `json:"-"`
	RelatedIssue uint64 `json:"relatedIssues"`
	Comment      string `json:"comment"`
	ProjectID    int64  `json:"projectId"`
}

// Check 检查请求参数是否合法
func (irc *IssueRelationCreateRequest) Check() error {
	if irc.IssueID == 0 {
		return errors.New("issueId is required")
	}

	if irc.RelatedIssue == 0 {
		return errors.New("relatedIssue is required")
	}

	if irc.ProjectID == 0 {
		return errors.New("projectId is required")
	}

	return nil
}

// IssueRelationGetResponse 事件关联关系响应
type IssueRelationGetResponse struct {
	Header
	UserInfoHeader
	Data *IssueRelations `json:"data"`
}

// IssueRelations 事件关联关系
type IssueRelations struct {
	RelatingIssues []Issue
	RelatedIssues  []Issue
}
