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
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
)

func (svc *Service) batchConvertTestCases(projectID uint64, models []dao.TestCase) ([]*apistructs.TestCase, error) {
	if len(models) == 0 {
		return nil, nil
	}
	// batch query api infos
	var tcIDs []uint64
	for _, model := range models {
		tcIDs = append(tcIDs, model.ID)
	}
	apis, err := svc.BatchListAPIs(projectID, tcIDs)
	if err != nil {
		return nil, err
	}

	// batch convert models
	var tcs []*apistructs.TestCase
	for _, model := range models {
		// apiCount
		var apiCount apistructs.TestCaseAPICount
		for _, api := range apis[model.ID] {
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
			APIs:           apis[model.ID],
			APICount:       apiCount,
			CreatedAt:      model.CreatedAt,
			UpdatedAt:      model.UpdatedAt,
		}
		tcs = append(tcs, &testCase)
	}

	return tcs, nil
}

// convertTestCase
// consider if you can use batchConvertTestCases firstly.
func (svc *Service) convertTestCase(model dao.TestCase) (*apistructs.TestCase, error) {
	tcs, err := svc.batchConvertTestCases(model.ProjectID, []dao.TestCase{model})
	if err != nil {
		return nil, err
	}
	if len(tcs) == 0 {
		return nil, fmt.Errorf("not found model")
	}
	return tcs[0], nil
}
