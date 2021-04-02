package model

type ManualReview struct {
	BaseModel
	BuildId         int    `gorm:"build_id"`
	ProjectId       int    `gorm:"project_id"`
	ApplicationId   int    `gorm:"application_id"`
	ApplicationName string `gorm:"application_name"`
	SponsorId       string `gorm:"sponsor_id"`
	CommitID        string `gorm:"commit_id"`
	OrgId           int64  `gorm:"org_id"`
	TaskId          int    `gorm:"task_id"`
	ProjectName     string `gorm:"project_name"`
	BranchName      string `gorm:"branch_name"`
	ApprovalStatus  string `gorm:"approval_status"`
	CommitMessage   string `gorm:"commit_message"`
	ApprovalReason  string `gorm:"approval_reason"`
}

type ReviewUser struct {
	BaseModel
	Operator string `gorm:"operator"`
	OrgId    int64  `gorm:"org_id"`
	TaskId   int64  `gorm:"task_id"`
}
