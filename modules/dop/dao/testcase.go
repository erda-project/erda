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
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// TestCase 测试用例
type TestCase struct {
	dbengine.BaseModel
	Name           string
	ProjectID      uint64
	TestSetID      uint64
	Priority       apistructs.TestCasePriority
	PreCondition   string
	StepAndResults TestCaseStepAndResults
	Desc           string
	Recycled       *bool
	From           apistructs.TestCaseFrom
	CreatorID      string
	UpdaterID      string
}

type TestCaseStepAndResults []apistructs.TestCaseStepAndResult

// TableName 设置模型对应数据库表名称
func (TestCase) TableName() string {
	return "dice_test_cases"
}

func (sr TestCaseStepAndResults) Value() (driver.Value, error) {
	if b, err := json.Marshal(sr); err != nil {
		return nil, errors.Errorf("failed to marshal stepAndResults, err: %v", err)
	} else {
		return string(b), nil
	}
}
func (sr *TestCaseStepAndResults) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return errors.New("invalid scan source for stepAndResults")
	}
	if len(v) == 0 {
		return nil
	}
	if err := json.Unmarshal(v, sr); err != nil {
		return errors.Wrapf(err, "failed to unmarshal stepAndResults")
	}
	return nil
}

// CreateTestCase 创建测试用例
func (client *DBClient) CreateTestCase(uc *TestCase) error {
	return client.Create(uc).Error
}

// UpdateTestCase 更新测试用例
func (client *DBClient) UpdateTestCase(uc *TestCase) error {
	return client.Save(uc).Error
}

func (client *DBClient) GetTestCaseByID(id uint64) (*TestCase, error) {
	var tc TestCase
	err := client.First(&tc, id).Error
	return &tc, err
}

func (client *DBClient) ListTestCasesByIDs(ids []uint64) ([]TestCase, error) {
	var tcs []TestCase
	if err := client.Where("`id` IN (?)", ids).Find(&tcs).Error; err != nil {
		return nil, err
	}
	if len(tcs) != len(ids) {
		m := make(map[uint64]struct{})
		for _, tc := range tcs {
			m[uint64(tc.ID)] = struct{}{}
		}
		var notFoundIDs []uint64
		for _, id := range ids {
			if _, ok := m[id]; !ok {
				notFoundIDs = append(notFoundIDs, id)
			}
		}
		if len(notFoundIDs) > 0 {
			return nil, fmt.Errorf("some test case not found, ids: %v", notFoundIDs)
		}
	}
	return tcs, nil
}

func (client *DBClient) ListTestCasesByTestSetIDs(req apistructs.TestCaseListRequest) ([]TestCase, error) {
	sql := client.DB
	if req.ProjectID > 0 {
		sql = sql.Where("`project_id` = ?", req.ProjectID)
	}
	if len(req.IDs) > 0 {
		sql = sql.Where("`id` IN (?)", req.IDs)
	}
	if len(req.TestSetIDs) > 0 {
		sql = sql.Where("`test_set_id` IN (?)", req.TestSetIDs)
	}
	if req.IDOnly {
		sql = sql.Select("`id`")
	}
	var tcs []TestCase
	if err := sql.Find(&tcs).Error; err != nil {
		return nil, err
	}
	return tcs, nil
}

func (client *DBClient) BatchUpdateTestCases(req apistructs.TestCaseBatchUpdateRequest) error {
	if len(req.TestCaseIDs) == 0 {
		return fmt.Errorf("no testcase selected")
	}

	sql := client.Model(TestCase{}).Where("`id` IN (?)", req.TestCaseIDs)

	kvs := make(map[string]interface{})

	if req.Priority != "" {
		kvs["priority"] = req.Priority
	}
	if req.Recycled != nil {
		kvs["recycled"] = req.Recycled
	}
	if req.MoveToTestSetID != nil {
		kvs["test_set_id"] = *req.MoveToTestSetID
	}

	return sql.Updates(kvs).Error
}

func (client *DBClient) BatchCopyTestCases(req apistructs.TestCaseBatchCopyRequest) error {
	if len(req.TestCaseIDs) == 0 {
		return fmt.Errorf("no testcase selected")
	}
	return nil
}

// RecycledTestCasesByTestSetID 回收测试集下的测试用例
func (client *DBClient) RecycledTestCasesByTestSetID(projectID, testSetID uint64) error {
	var useCase TestCase
	return client.Model(&useCase).
		Where("project_id = ?", projectID).
		Where("test_set_id = ?", testSetID).Updates("recycled", apistructs.RecycledYes).Error
}

// RecoverTestCasesByTestSetID 回收站恢复测试集下的测试用例
func (client *DBClient) RecoverTestCasesByTestSetID(projectID, testSetID uint64) error {
	var useCase TestCase
	return client.Model(&useCase).
		Where("project_id = ?", projectID).
		Where("test_set_id = ?", testSetID).Updates("recycled", apistructs.RecycledNo).Error
}

// CleanTestCasesByTestSetID 彻底删除测试集下的测试用例
func (client *DBClient) CleanTestCasesByTestSetID(projectID, testSetID uint64) error {
	return client.
		Where("project_id = ?", projectID).
		Where("test_set_id = ?", testSetID).
		Delete(TestCase{}).Error
}

func (client *DBClient) BatchDeleteTestCases(ids []uint64) error {
	return client.Where("`id` IN (?)", ids).Delete(TestCase{}).Error
}
