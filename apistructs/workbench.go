package apistructs


type WorkbenchRequest struct {
	OrgID uint64 `json:"orgID"`
	PageNo int `json:"pageNo"`
	PageSize int `json:"pageSize"`
	IssueSize int `json:"issueSize"`
}

type WorkbenchResponse struct {
	Header
	Data WorkbenchResponseData `json:"data"`
}

type WorkbenchResponseData struct {
	Total int `json:"total"`
	TotalIssue int `json:"totalIssue"`
	List []WorkbenchProjectItem `json:"list"`
}

type WorkbenchProjectItem struct {
	ProjectDTO ProjectDTO `json:"projectDTO"`
	TotalIssueNum int `json:"totalIssueNum"`
	ExpiredIssueNum int `json:"expiredIssueNum"`
	ExpiredOneDayNum int `json:"expiredOneDayNum"`
	ExpiredSevenDayNum int `json:"expiredSevenDayNum"`
	ExpiredThirtyDayNum int `json:"expiredThirtyDayNum"`
	FeatureDayNum int `json:"featureDayNum"`
	IssueList []Issue `json:"issueList"`
}