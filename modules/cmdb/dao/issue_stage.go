package dao

import (
	"github.com/erda-project/erda/apistructs"
)

type IssueStage struct {
	BaseModel
	OrgID     int64                `gorm:"column:org_id"`
	IssueType apistructs.IssueType `gorm:"column:issue_type"`
	Name      string               `gorm:"column:name"`
	Value     string               `gorm:"column:value"`
}

func (i IssueStage) TableName() string {
	return "dice_issue_stage"
}

func (client *DBClient) CreateIssueStage(stages []IssueStage) error {
	return client.BulkInsert(stages)
}

func (client *DBClient) DeleteIssuesStage(orgID int64, issueType apistructs.IssueType) error {
	return client.Table("dice_issue_stage").Where("org_id = ?", orgID).
		Where("issue_type = ?", issueType).Delete(IssueStage{}).Error
}

func (client *DBClient) GetIssuesStage(orgID int64, issueType apistructs.IssueType) ([]IssueStage, error) {
	var stages []IssueStage
	err := client.Table("dice_issue_stage").Where("org_id = ?", orgID).
		Where("issue_type = ?", issueType).Find(&stages).Error
	if err != nil {
		return nil, err
	}
	return stages, nil
}
