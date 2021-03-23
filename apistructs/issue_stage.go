package apistructs

type IssueStageRequest struct {
	OrgID     int64        `json:"orgID"`
	IssueType IssueType    `json:"issueType"`
	List      []IssueStage `json:"list"`
	IdentityInfo
}

type IssueStage struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type IssueStageResponse struct {
	Header
	Data []IssueStage `json:"data"`
}
