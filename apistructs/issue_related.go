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

package apistructs

import "github.com/pkg/errors"

type IssueRelationCreateRequest struct {
	IssueID      uint64 `json:"-"`
	RelatedIssue uint64 `json:"relatedIssues"`
	Comment      string `json:"comment"`
	ProjectID    int64  `json:"projectId"`
	Type         string `json:"type"`
}

const IssueRelationConnection = "connection"
const IssueRelationInclusion = "inclusion"

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

	if len(irc.Type) == 0 {
		return errors.New("type is required")
	}

	if irc.Type != IssueRelationConnection && irc.Type != IssueRelationInclusion {
		return errors.New("invalid issue relation type")
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

type IssueRelationRequest struct {
	RelationTypes []string `schema:"type"`
}
