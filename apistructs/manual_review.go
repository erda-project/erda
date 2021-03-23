package apistructs

import "time"

type GetReviewsBySponsorIdRequest struct {
	SponsorId      int64  `json:"sponsorId"`
	Id             int64  `json:"id"`
	ProjectId      int    `json:"projectId"`
	ApprovalStatus string `json:"approvalStatus"`
	OrgId          int64  `json:"orgId"`
	PageNo         int    `json:"pageNo"`
	PageSize       int    `json:"pageSize"`
	IdentityInfo
}
type GetReviewsBySponsorIdResponse struct {
	Id              int64    `json:"id"`
	ProjectId       int      `json:"projectId"`
	ProjectName     string   `json:"projectName"`
	ApplicationId   int      `json:"applicationId"`
	ApplicationName string   `json:"applicationName"`
	BuildId         int      `json:"buildId"`
	BranchName      string   `json:"branchName"`
	CommitId        string   `json:"commitId"`
	CommitMessage   string   `json:"commitMessage"`
	Approver        []string `json:"approver"`
	ApprovalContent string   `json:"approvalContent"`
	ApprovalReason  string   `json:"approvalReason"`
}

type CreateReviewRequest struct {
	BuildId         int       `json:"buildId"`
	ProjectId       int       `json:"projectId"`
	ApplicationId   int       `json:"applicationId"`
	ApplicationName string    `json:"applicationName"`
	SponsorId       string    `json:"sponsorId"`
	CommitID        string    `json:"commitID"`
	OrgId           int64     `json:"orgId"`
	TaskId          int       `json:"taskId"`
	ProjectName     string    `json:"projectName"`
	BranchName      string    `json:"branchName"`
	ApprovalStatus  string    `json:"approvalStatus"`
	CommitMessage   string    `json:"commitMessage"`
	CreatedAt       time.Time `json:"createdAt"`
}
type GetReviewsByUserIdRequest struct {
	UserId         int64  `json:"userId"`
	Id             int64  `json:"id"`
	ProjectId      int    `json:"buildId"`
	Operator       int64  `json:"operator"`
	ApprovalStatus string `json:"approvalStatus"`
	OrgId          int64  `json:"orgId"`
	PageNo         int    `json:"pageNo"`
	PageSize       int    `json:"pageSize"`
}

type GetAuthorityByUserIdRequest struct {
	Operator int64 `json:"operator"`
	OrgId    int64 `json:"orgId"`
	TaskId   int64 `json:"TaskId"`
}

type GetAuthorityByUserIdResponse struct {
	Authority string `json:"authority"`
}

type GetReviewByTaskIdIdRequest struct {
	TaskId int64 `json:"TaskId"`
}

type GetReviewByTaskIdIdResponse struct {
	Total          int    `json:"total"`
	ApprovalStatus string `json:"approvalStatus"`
	Id             int64  `json:"id"`
}

type CreateReviewUser struct {
	OrgId     int64     `json:"orgId"`
	Operator  string    `json:"operator"`
	TaskId    int64     `json:"taskId"`
	CreatedAt time.Time `json:"createdAt"`
}
type GetReviewsByUserIdResponse struct {
	Id              int64  `json:"id"`
	ProjectName     string `json:"projectName"`
	ApplicationName string `json:"applicationName"`
	ProjectId       int    `json:"projectId"`
	ApplicationId   int    `json:"applicationId"`
	BuildId         int    `json:"buildId"`
	BranchName      string `json:"branchName"`
	CommitId        string `json:"commitId"`
	CommitMessage   string `json:"commitMessage"`
	Operator        string `json:"operator"`
	ApprovalStatus  string `json:"approvalStatus"`
	ApprovalContent string `json:"approvalContent"`
	ApprovalReason  string `json:"approvalReason"`
}

type UpdateApproval struct {
	Id     int64  `json:"id"`
	OrgId  int64  `json:"orgId"`
	Reject bool   `json:"reject"`
	Reason string `json:"reason" gorm:"commit_message"`
}

type ReviewsBySponsorList struct {
	List  []GetReviewsBySponsorIdResponse `json:"list"`
	Total int                             `json:"total"`
}

type ReviewsByUserList struct {
	List  []GetReviewsByUserIdResponse `json:"list"`
	Total int                          `json:"total"`
}
