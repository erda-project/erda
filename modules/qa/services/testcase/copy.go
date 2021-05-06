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

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (svc *Service) BatchCopyTestCases(req apistructs.TestCaseBatchCopyRequest) ([]uint64, error) {
	// 校验目标测试集是否存在
	if req.CopyToTestSetID > 0 {
		if _, err := svc.db.GetTestSetByID(req.CopyToTestSetID); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return nil, apierrors.ErrBatchCopyTestCases.InvalidParameter(fmt.Sprintf("testset not found, id: %d", req.CopyToTestSetID))
			}
			return nil, apierrors.ErrBatchCopyTestCases.InvalidParameter(fmt.Sprintf("failed to find testset, id: %d, err: %v", req.CopyToTestSetID, err))
		}
	}

	// 校验项目 ID
	if req.ProjectID == 0 {
		return nil, apierrors.ErrBatchCopyTestCases.MissingParameter("projectID")
	}

	// 校验 ids 是否都存在
	if len(req.TestCaseIDs) == 0 {
		return nil, apierrors.ErrBatchCopyTestCases.MissingParameter("testCaseIDs")
	}
	fromTestCases, err := svc.db.ListTestCasesByIDs(req.TestCaseIDs)
	if err != nil {
		return nil, apierrors.ErrBatchCopyTestCases.InvalidParameter(err)
	}

	// 批量复制 -> 批量创建
	batchCreateReq := apistructs.TestCaseBatchCreateRequest{
		ProjectID:    req.ProjectID,
		TestCases:    nil,
		IdentityInfo: req.IdentityInfo,
	}
	for _, dbTc := range fromTestCases {
		//// 若目标测试集和当前测试集相同，则不复制
		//if dbTc.TestSetID == req.CopyToTestSetID {
		//	continue
		//}
		tc, err := svc.convertTestCase(dbTc)
		if err != nil {
			return nil, err
		}
		batchCreateReq.TestCases = append(batchCreateReq.TestCases, apistructs.TestCaseCreateRequest{
			ProjectID:      req.ProjectID,
			TestSetID:      req.CopyToTestSetID,
			Name:           tc.Name,
			PreCondition:   tc.PreCondition,
			StepAndResults: tc.StepAndResults,
			APIs:           tc.APIs,
			Desc:           tc.Desc,
			Priority:       tc.Priority,
			LabelIDs:       tc.LabelIDs,
			IdentityInfo:   req.IdentityInfo,
		})
	}
	return svc.BatchCreateTestCases(batchCreateReq)
}
