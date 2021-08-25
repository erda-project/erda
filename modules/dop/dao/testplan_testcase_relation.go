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

package dao

import (
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/strutil"
)

// TestPlanCaseRel
type TestPlanCaseRel struct {
	dbengine.BaseModel
	TestPlanID uint64
	TestSetID  uint64
	TestCaseID uint64
	ExecStatus apistructs.TestCaseExecStatus
	CreatorID  string
	UpdaterID  string
	ExecutorID string
}

// TableName 表名
func (TestPlanCaseRel) TableName() string {
	return "dice_test_plan_case_relations"
}

// CreateTestPlanCaseRel Create testPlanCaseRel
func (client *DBClient) CreateTestPlanCaseRel(testPlanCaseRel *TestPlanCaseRel) error {
	return client.Create(testPlanCaseRel).Error
}

func (client *DBClient) BatchCreateTestPlanCaseRels(rels []TestPlanCaseRel) error {
	return client.BulkInsert(rels)
}

func (client *DBClient) GetTestPlanCaseRelByPlanIDAndCaseID(planID, caseID uint64) (*TestPlanCaseRel, error) {
	var rel TestPlanCaseRel
	if err := client.Where("`test_plan_id` = ?", planID).Where("`test_case_id` = ?", caseID).First(&rel).Error; err != nil {
		return nil, err
	}
	return &rel, nil
}

// CheckTestPlanCaseRelIDsExistOrNot 返回已存在和不存在的 id 列表
// return: existIDs, notExistIDs, error
func (client *DBClient) CheckTestPlanCaseRelIDsExistOrNot(planID uint64, caseIDs []uint64) ([]uint64, []uint64, error) {
	var existRels []TestPlanCaseRel
	if err := client.Select("`test_case_id`").Where("`test_plan_id` = ?", planID).
		Where("`test_case_id` IN (?)", caseIDs).Find(&existRels).Error; err != nil {
		return nil, nil, err
	}
	var existIDs []uint64
	for _, rel := range existRels {
		existIDs = append(existIDs, rel.TestCaseID)
	}
	allIDMap := make(map[uint64]struct{})
	for _, id := range caseIDs {
		allIDMap[id] = struct{}{}
	}
	for _, existID := range existIDs {
		delete(allIDMap, existID)
	}
	var notExistIDs []uint64
	for id := range allIDMap {
		notExistIDs = append(notExistIDs, id)
	}
	return existIDs, notExistIDs, nil
}

// UpdateTestPlanTestCaseRel Update testPlanCaseRel
func (client *DBClient) UpdateTestPlanTestCaseRel(testPlanCaseRel *TestPlanCaseRel) error {
	return client.Save(testPlanCaseRel).Error
}

// BatchUpdateTestPlanCaseRels 批量更新测试计划用例
func (client *DBClient) BatchUpdateTestPlanCaseRels(req apistructs.TestPlanCaseRelBatchUpdateRequest) error {
	db := client.Model(TestPlanCaseRel{}).Where("`test_plan_id` = ?", req.TestPlanID)
	if len(req.RelationIDs) > 0 {
		db = db.Where("`id` IN (?)", req.RelationIDs)
	}
	updateFields := make(map[string]interface{})
	if req.ExecutorID != "" {
		updateFields["executor_id"] = req.ExecutorID
	}
	if req.ExecStatus != "" {
		updateFields["exec_status"] = req.ExecStatus
	}
	if req.IdentityInfo.UserID != "" {
		updateFields["updater_id"] = req.IdentityInfo.UserID
	}
	return db.Updates(updateFields).Error
}

// DeleteTestPlanCaseRelations
func (client *DBClient) DeleteTestPlanCaseRelations(testPlanID uint64, relIDs []uint64) error {
	sql := client.Where("test_plan_id = ?", testPlanID)
	if len(relIDs) > 0 {
		sql = sql.Where("id IN (?)", relIDs)
	}
	return sql.Delete(&TestPlanCaseRel{}).Error
}

// DeleteTestPlanCaseRelationsByTestCaseIds
func (client *DBClient) DeleteTestPlanCaseRelationsByTestCaseIds(testCaseIds []uint64) error {
	sql := client.Where("test_case_id IN (?)", testCaseIds)
	return sql.Delete(&TestPlanCaseRel{}).Error
}

// DeleteTestPlanCaseRelByTestPlanID Delete relations by testPlanID
func (client *DBClient) DeleteTestPlanCaseRelByTestPlanID(testPlanID uint64) error {
	return client.Where("testplan_id = ?", testPlanID).Delete(TestPlanCaseRel{}).Error
}

// DeleteTestPlanCaseRelByTestCaseID Delete relations by testCaseID
func (client *DBClient) DeleteTestPlanCaseRelByTestCaseID(testCaseID uint64) error {
	return client.Where("usecase_id = ?", testCaseID).Delete(TestPlanCaseRel{}).Error
}

// GetTestPlanCaseRel Fetch testPlanCaseRel
func (client *DBClient) GetTestPlanCaseRel(relID uint64) (*TestPlanCaseRel, error) {
	var testPlanCaseRel TestPlanCaseRel
	if err := client.Where("id = ?", relID).First(&testPlanCaseRel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &testPlanCaseRel, nil
}

// ListTestPlanCaseRels List testPlanCaseRel
func (client *DBClient) ListTestPlanCaseRels(req apistructs.TestPlanCaseRelListRequest) ([]TestPlanCaseRel, error) {
	var rels []TestPlanCaseRel
	sql := client.DB
	if len(req.IDs) > 0 {
		sql = sql.Where("`id` IN (?)", req.IDs)
	}
	if len(req.TestPlanIDs) > 0 {
		sql = sql.Where("`test_plan_id` IN (?)", req.TestPlanIDs)
	}
	if len(req.TestSetIDs) > 0 {
		sql = sql.Where("`test_set_id` IN (?)", req.TestSetIDs)
	}
	if len(req.CreatorIDs) > 0 {
		sql = sql.Where("`creator_id` IN (?)", req.CreatorIDs)
	}
	if len(req.UpdaterIDs) > 0 {
		sql = sql.Where("`updater_id` IN (?)", req.UpdaterIDs)
	}
	if len(req.ExecutorIDs) > 0 {
		if strutil.Exist(req.ExecutorIDs, "unassigned") {
			sql = sql.Where("`executor_id` IN (?) OR `executor_id` = ''", req.ExecutorIDs)
		} else {
			sql = sql.Where("`executor_id` IN (?)", req.ExecutorIDs)
		}
	}
	if len(req.ExecStatuses) > 0 {
		sql = sql.Where("`exec_status` IN (?)", req.ExecStatuses)
	}
	if req.UpdatedAtBeginInclude != nil {
		sql = sql.Where("`updated_at` >= ?", req.UpdatedAtBeginInclude)
	}
	if req.UpdatedAtEndInclude != nil {
		sql = sql.Where("`updated_at` <= ?", req.UpdatedAtEndInclude)
	}
	if req.IDOnly {
		sql = sql.Select("`id`")
	}
	sql = sql.Order("`id` DESC")
	if err := sql.Find(&rels).Error; err != nil {
		return nil, err
	}

	return rels, nil
}

// ListTestPlanTestSetIDs 获取测试计划下的测试集 ID 列表，从关联关系而来
func (client *DBClient) ListTestPlanTestSetIDs(testPlanID uint64) ([]uint64, error) {
	var rels []TestPlanCaseRel
	sql := client.DB.Where("`test_plan_id` = ?", testPlanID).Find(&rels)
	if err := sql.Error; err != nil {
		return nil, err
	}
	testSetIDMap := make(map[uint64]struct{})
	for _, rel := range rels {
		testSetIDMap[rel.TestSetID] = struct{}{}
	}
	var testSetIDs []uint64
	for tsID := range testSetIDMap {
		testSetIDs = append(testSetIDs, tsID)
	}
	return testSetIDs, nil
}

func (client *DBClient) ListTestPlanCaseRelsCount(testPlanIDs []uint64) (map[uint64]apistructs.TestPlanRelsCount, error) {
	if len(testPlanIDs) == 0 {
		return nil, nil
	}

	// temp is a struct for `select test_plan_id, exec_status, count(*) as count from ...`
	type temp struct {
		TestPlanCaseRel
		Count uint64
	}
	var rels []temp
	sql := client.Table(TestPlanCaseRel{}.TableName()).
		Select([]string{"`test_plan_id`", "`exec_status`", "count(*) as count"}).
		Where("`test_plan_id` IN (?)", testPlanIDs).
		Group("`test_plan_id`, `exec_status`").Find(&rels)
	if err := sql.Error; err != nil {
		return nil, err
	}

	result := make(map[uint64]apistructs.TestPlanRelsCount)
	for _, rel := range rels {
		c, ok := result[rel.TestPlanID]
		if !ok {
			c = apistructs.TestPlanRelsCount{}
		}
		switch rel.ExecStatus {
		case apistructs.CaseExecStatusInit:
			c.Init = rel.Count
		case apistructs.CaseExecStatusSucc:
			c.Succ = rel.Count
		case apistructs.CaseExecStatusFail:
			c.Fail = rel.Count
		case apistructs.CaseExecStatusBlocked:
			c.Block = rel.Count
		}
		c.Total += rel.Count
		result[rel.TestPlanID] = c
	}

	return result, nil
}
