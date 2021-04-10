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
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

// BatchCleanFromRecycleBin 从回收站彻底删除测试用例
func (svc *Service) BatchCleanFromRecycleBin(req apistructs.TestCaseBatchCleanFromRecycleBinRequest) error {
	// 确认所有 testCaseIDs 存在
	tcs, err := svc.db.ListTestCasesByIDs(req.TestCaseIDs)
	if err != nil {
		return apierrors.ErrBatchCleanTestCasesFromRecycleBin.InvalidParameter(err)
	}

	// 校验 recycled
	var notRecycledIDs []uint64
	for _, tc := range tcs {
		if !*tc.Recycled {
			notRecycledIDs = append(notRecycledIDs, uint64(tc.ID))
		}
	}
	if len(notRecycledIDs) > 0 {
		return apierrors.ErrBatchCleanTestCasesFromRecycleBin.InvalidParameter(fmt.Errorf("some testcase not in recycle bin, ids: %v", notRecycledIDs))
	}

	// 批量删除缺陷测试计划用例关联
	if err := svc.db.DeleteIssueTestCaseRelationsByTestCaseIDs(req.TestCaseIDs); err != nil {
		return apierrors.ErrBatchCleanTestCasesFromRecycleBin.InternalError(err)
	}

	// 批量删除测试用例
	if err := svc.db.BatchDeleteTestCases(req.TestCaseIDs); err != nil {
		return apierrors.ErrBatchCleanTestCasesFromRecycleBin.InternalError(err)
	}

	return nil
}
