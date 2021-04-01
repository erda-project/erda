package dao

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
)

// IssueTestCaseRelation 事件与用例关联
type IssueTestCaseRelation struct {
	BaseModel
	IssueID           uint64 `json:"issueID"`
	TestPlanID        uint64 `json:"testPlanID"`
	TestPlanCaseRelID uint64 `json:"testPlanCaseRelID"`
	TestCaseID        uint64 `json:"testCaseID"`
	CreatorID         string `json:"creatorID"`
}

// TableName 表名
func (IssueTestCaseRelation) TableName() string {
	return "dice_issue_testcase_relations"
}

// ListIssueTestCaseRelations 查询事件用例关联关系列表
func (client *DBClient) ListIssueTestCaseRelations(req apistructs.IssueTestCaseRelationsListRequest) ([]IssueTestCaseRelation, error) {
	// 参数校验
	if req.IssueID == 0 && req.TestPlanID == 0 && req.TestPlanCaseRelID == 0 && req.TestCaseID == 0 {
		return nil, fmt.Errorf("empty request")
	}

	sql := client.DB
	if req.IssueID > 0 {
		sql = sql.Where("`issue_id` = ?", req.IssueID)
	}
	if req.TestPlanID > 0 {
		sql = sql.Where("`test_plan_id` = ?", req.TestPlanID)
	}
	if req.TestPlanCaseRelID > 0 {
		sql = sql.Where("`test_plan_case_rel_id` = ?", req.TestPlanCaseRelID)
	}
	if req.TestCaseID > 0 {
		sql = sql.Where("`test_case_id` = ?", req.TestCaseID)
	}
	var results []IssueTestCaseRelation
	if err := sql.Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// BatchCreateIssueTestCaseRelations 批量创建关联关系
func (client *DBClient) BatchCreateIssueTestCaseRelations(rels []IssueTestCaseRelation) error {
	return client.BulkInsert(rels)
}

// DeleteIssueTestCaseRelationsByIssueID 根据 issue id 删除关联关系
func (client *DBClient) DeleteIssueTestCaseRelationsByIssueID(issueID uint64) error {
	return client.Where("`issue_id` = ?", issueID).Delete(&IssueTestCaseRelation{}).Error
}

// DeleteIssueTestCaseRelationsByIDs 根据关联关系 id 删除
func (client *DBClient) DeleteIssueTestCaseRelationsByIDs(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	return client.Debug().Where("`id` IN (?)", ids).Delete(&IssueTestCaseRelation{}).Error
}
