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
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

// CreateTestCase 创建测试用例
func (svc *Service) CreateTestCase(req apistructs.TestCaseCreateRequest) (uint64, error) {
	// 参数检查
	if req.Name == "" {
		return 0, apierrors.ErrCreateTestCase.MissingParameter("name")
	}
	if req.Priority == "" {
		return 0, apierrors.ErrCreateTestCase.MissingParameter("priority")
	}
	if !req.Priority.IsValid() {
		return 0, apierrors.ErrCreateTestCase.InvalidParameter(fmt.Sprintf("priority: %s", req.Priority))
	}

	tc := dao.TestCase{
		Name:           req.Name,
		StepAndResults: dao.TestCaseStepAndResults(req.StepAndResults),
		From:           apistructs.TestCaseFromManual,
		ProjectID:      req.ProjectID,
		CreatorID:      req.IdentityInfo.UserID,
		UpdaterID:      req.IdentityInfo.UserID,
		PreCondition:   req.PreCondition,
		Recycled:       &[]bool{false}[0],
		Desc:           req.Desc,
		TestSetID:      req.TestSetID,
		Priority:       req.Priority,
	}
	if err := svc.db.CreateTestCase(&tc); err != nil {
		return 0, apierrors.ErrCreateTestCase.InternalError(fmt.Errorf("failed to insert testcase into database, err: %v", err))
	}

	// 创建 API 信息
	if len(req.APIs) > 0 {
		if err := svc.createOrUpdateAPIs(uint64(tc.ID), req.ProjectID, req.APIs); err != nil {
			return 0, apierrors.ErrCreateTestCase.InternalError(fmt.Errorf("failed to create api info, err: %v", err))
		}
	}

	return uint64(tc.ID), nil
}

// BatchCreateTestCases 批量创建测试用例
func (svc *Service) BatchCreateTestCases(req apistructs.TestCaseBatchCreateRequest) ([]uint64, error) {
	// 参数校验
	if req.ProjectID == 0 {
		return nil, apierrors.ErrBatchCreateTestCases.MissingParameter("projectID")
	}

	// 遍历创建
	if len(req.TestCases) == 0 {
		return nil, nil
	}

	var allCreatedTestCaseIDs []uint64

	for _, tcReq := range req.TestCases {
		// pre handle
		tcReq.ProjectID = req.ProjectID
		tcReq.IdentityInfo = req.IdentityInfo
		// handle apis
		for i := range tcReq.APIs {
			tcReq.APIs[i].ApiID = 0
		}

		tcID, err := svc.CreateTestCase(tcReq)
		if err != nil {
			return nil, err
		}

		allCreatedTestCaseIDs = append(allCreatedTestCaseIDs, tcID)
	}

	return allCreatedTestCaseIDs, nil
}
