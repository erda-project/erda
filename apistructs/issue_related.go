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
