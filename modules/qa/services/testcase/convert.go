// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package testcase

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

// convertTestCase 将 model 转换为 apistruct
func (svc *Service) convertTestCase(model dao.TestCase) (*apistructs.TestCase, error) {
	// 获取 API 信息
	apiTests, err := svc.GetApiTestListByUsecaseID(int64(model.ID))
	if err != nil {
		return nil, apierrors.ErrGetApiTestInfo.InternalError(err)
	}
	// apiCount
	var apiCount apistructs.TestCaseAPICount
	for _, api := range apiTests {
		apiCount.Total++
		switch api.Status {
		case apistructs.ApiTestCreated:
			apiCount.Created++
		case apistructs.ApiTestRunning:
			apiCount.Running++
		case apistructs.ApiTestPassed:
			apiCount.Passed++
		case apistructs.ApiTestFailed:
			apiCount.Failed++
		}
	}

	// convert
	testCase := apistructs.TestCase{
		ID:             uint64(model.ID),
		Name:           model.Name,
		Priority:       model.Priority,
		PreCondition:   model.PreCondition,
		Desc:           model.Desc,
		Recycled:       model.Recycled,
		TestSetID:      model.TestSetID,
		ProjectID:      model.ProjectID,
		CreatorID:      model.CreatorID,
		UpdaterID:      model.UpdaterID,
		BugIDs:         nil,
		LabelIDs:       nil,
		Attachments:    nil,
		StepAndResults: model.StepAndResults,
		Labels:         nil,
		APIs:           apiTests,
		APICount:       apiCount,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}

	return &testCase, nil
}
