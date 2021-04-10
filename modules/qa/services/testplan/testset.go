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

package testplan

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

func (t *TestPlan) ListTestSet(req apistructs.TestPlanTestSetsListRequest) ([]apistructs.TestSet, error) {
	// 参数校验
	if req.TestPlanID == 0 {
		return nil, apierrors.ErrListTestPlanTestSets.MissingParameter("testPlanID")
	}

	tsIDs, err := t.db.ListTestPlanTestSetIDs(req.TestPlanID)
	if err != nil {
		return nil, apierrors.ErrListTestPlanTestSets.InternalError(err)
	}
	if len(tsIDs) == 0 {
		return nil, nil
	}

	return t.testSetSvc.ListTestSetByLeafTestSetIDs(req.ParentTestSetID, tsIDs)
}
