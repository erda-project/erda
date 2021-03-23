package apistructs

type IssuePanel struct {
	PanelName string `json:"panelName"`
	PanelID   int64  `json:"panelID"`
}

type IssuePanelIssues struct {
	IssuePanel
	Total int64 `json:"total"`
}

type IssuePanelIssueIDs struct {
	Issues []Issue `json:"issues"`
	Total  uint64  `json:"total"`
}

// 自定义看板请求
type IssuePanelRequest struct {
	IssuePanel
	IssueID int64 `json:"issueID"`
	IssuePagingRequest
	IdentityInfo
}

// 自定义看板创建响应
type IssuePanelIssuesCreateResponse struct {
	Header
	Data int64 `json:"data"`
}

// 自定义看板查询响应
type IssuePanelGetResponse struct {
	Header
	Data []IssuePanelIssues `json:"data"`
}

// 查询自定义看板内事件响应
type IssuePanelIssuesGetResponse struct {
	Header
	Data *IssuePanelIssueIDs `json:"data"`
}

// 自定义看板删除响应
type IssuePanelDeleteResponse struct {
	Header
	Data *IssuePanel `json:"data"`
}
