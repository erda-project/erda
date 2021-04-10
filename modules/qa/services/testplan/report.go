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
	"github.com/erda-project/erda/pkg/strutil"
)

func (t *TestPlan) GenerateReport(testPlanID uint64) (*apistructs.TestPlanReport, error) {
	var report apistructs.TestPlanReport

	// 查询测试计划
	tp, err := t.Get(testPlanID)
	if err != nil {
		return nil, err
	}
	report.TestPlan = *tp

	// 查询关联关系列表
	relsCountMap, err := t.db.ListTestPlanCaseRelsCount([]uint64{testPlanID})
	if err != nil {
		return nil, apierrors.ErrPagingTestPlanCaseRels.InternalError(err)
	}
	report.RelsCount = relsCountMap[testPlanID]

	// 接口总数
	rels, err := t.db.ListTestPlanCaseRels(apistructs.TestPlanCaseRelListRequest{
		TestPlanIDs: []uint64{testPlanID},
	})
	if err != nil {
		return nil, apierrors.ErrPagingTestPlanCaseRels.InternalError(err)
	}
	var totalApiCount apistructs.TestCaseAPICount
	for _, rel := range rels {
		apis, err := t.testCaseSvc.ListAPIs(int64(rel.TestCaseID))
		if err != nil {
			return nil, apierrors.ErrGetApiTestInfo.InternalError(err)
		}
		for _, api := range apis {
			totalApiCount.Total++
			switch api.Status {
			case apistructs.ApiTestCreated:
				totalApiCount.Created++
			case apistructs.ApiTestRunning:
				totalApiCount.Running++
			case apistructs.ApiTestPassed:
				totalApiCount.Passed++
			case apistructs.ApiTestFailed:
				totalApiCount.Failed++
			}
		}
	}
	report.APICount = totalApiCount

	// 执行者的所属用例执行情况
	executorStatus := make(map[string]apistructs.TestPlanRelsCount)
	for _, rel := range rels {
		c := executorStatus[rel.ExecutorID]
		c.Total++
		switch rel.ExecStatus {
		case apistructs.CaseExecStatusInit:
			c.Init++
		case apistructs.CaseExecStatusSucc:
			c.Succ++
		case apistructs.CaseExecStatusFail:
			c.Fail++
		case apistructs.CaseExecStatusBlocked:
			c.Block++
		}
		executorStatus[rel.ExecutorID] = c
	}
	report.ExecutorStatus = executorStatus

	// userIDs
	var userIDs []string
	userIDs = append(append(userIDs, report.TestPlan.OwnerID, report.TestPlan.CreatorID, report.TestPlan.UpdaterID), report.TestPlan.PartnerIDs...)
	for executorID := range report.ExecutorStatus {
		userIDs = append(userIDs, executorID)
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	report.UserIDs = userIDs

	return &report, nil
}
