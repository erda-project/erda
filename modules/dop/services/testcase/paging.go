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
	"sort"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

// ListTestCases return result TestCases, TestCaseIDs and error if have.
func (svc *Service) ListTestCases(req apistructs.TestCaseListRequest) ([]apistructs.TestCase, []uint64, error) {
	if req.ProjectID == 0 && !req.AllowMissingProjectID {
		return nil, nil, apierrors.ErrListTestCases.MissingParameter("projectID")
	}
	if len(req.TestSetIDs) == 0 && !req.AllowEmptyTestSetIDs {
		return nil, nil, apierrors.ErrListTestCases.MissingParameter("testCaseIDs")
	}
	dbTcs, err := svc.db.ListTestCasesByTestSetIDs(req)
	if err != nil {
		return nil, nil, apierrors.ErrListTestCases.InternalError(err)
	}
	var tcs []apistructs.TestCase
	var tcsIDs []uint64
	for _, dbTc := range dbTcs {
		tc := apistructs.TestCase{ID: uint64(dbTc.ID)}
		if !req.IDOnly {
			converted, err := svc.convertTestCase(dbTc)
			if err != nil {
				return nil, nil, apierrors.ErrListTestCases.InternalError(fmt.Errorf("failed to convert testcase, id: %d, err: %v", tc.ID, err))
			}
			tc = *converted
		}
		tcs = append(tcs, tc)
		tcsIDs = append(tcsIDs, tc.ID)
	}
	return tcs, tcsIDs, nil
}

func (svc *Service) ListTestCasesDeep(req apistructs.TestCaseListRequest) ([]uint64, error) {
	_, resultTestCaseIDs, err := svc.ListTestCases(req)
	if err != nil {
		return nil, apierrors.ErrListTestCases.InternalError(err)
	}
	testSets := []dao.TestSet{}
	for _, testSetID := range req.TestSetIDs {
		testSets, err = svc.db.ListTestSets(apistructs.TestSetListRequest{
			ProjectID: &req.ProjectID,
			ParentID:  &testSetID,
		})
		if err != nil {
			return nil, apierrors.ErrListTestCases.InternalError(err)
		}
	}
	for {
		if len(testSets) == 0 {
			break
		}
		childTestSets := []dao.TestSet{}
		testSetIDs := []uint64{}
		for _, testSet := range testSets {
			testSetIDs = append(testSetIDs, testSet.ID)
			testSets, err := svc.db.ListTestSets(apistructs.TestSetListRequest{
				ProjectID: &req.ProjectID,
				ParentID:  &testSet.ID,
			})
			if err != nil {
				return nil, apierrors.ErrListTestCases.InternalError(err)
			}
			childTestSets = append(childTestSets, testSets...)
		}
		_, testCaseIDs, err := svc.ListTestCases(apistructs.TestCaseListRequest{
			ProjectID:            req.ProjectID,
			AllowEmptyTestSetIDs: false,
			TestSetIDs:           testSetIDs,
			Recycled:             false,
			IDOnly:               true,
		})
		if err != nil {
			return nil, apierrors.ErrListTestCases.InternalError(err)
		}
		resultTestCaseIDs = append(resultTestCaseIDs, testCaseIDs...)
		testSets = childTestSets
	}
	return resultTestCaseIDs, nil

}

// PagingTestCases ????????????????????????
func (svc *Service) PagingTestCases(req apistructs.TestCasePagingRequest) (*apistructs.TestCasePagingResponseData, error) {
	// order
	const (
		fieldPriority  = "priority"
		fieldID        = "id"
		fieldTestSetID = "test_set_id"
		fieldUpdaterID = "updater_id"
		fieldUpdatedAt = "updated_at"
	)
	// handle request
	if req.ProjectID == 0 {
		return nil, apierrors.ErrPagingTestCases.MissingParameter("projectID")
	}
	for _, priority := range req.Priorities {
		if !priority.IsValid() {
			return nil, apierrors.ErrPagingTestCases.InvalidParameter(fmt.Sprintf("priority: %s", priority))
		}
	}
	if req.OrderByPriorityAsc != nil && req.OrderByPriorityDesc != nil {
		return nil, apierrors.ErrPagingTestCases.InvalidParameter("order by priority ASC or DESC?")
	}
	if req.OrderByUpdaterIDAsc != nil && req.OrderByUpdaterIDDesc != nil {
		return nil, apierrors.ErrPagingTestCases.InvalidParameter("order by updaterID ASC or DESC?")
	}
	if req.OrderByUpdatedAtAsc != nil && req.OrderByUpdatedAtDesc != nil {
		return nil, apierrors.ErrPagingTestCases.InvalidParameter("order by updatedAt ASC or DESC?")
	}
	if req.OrderByIDAsc != nil && req.OrderByIDDesc != nil {
		return nil, apierrors.ErrPagingTestCases.InvalidParameter("order by id ASC or DESC?")
	}
	if req.OrderByTestSetIDAsc != nil && req.OrderByTestSetIDDesc != nil {
		return nil, apierrors.ErrPagingTestCases.InvalidParameter("order by testSetID ASC or DESC?")
	}
	if req.OrderByTestSetNameAsc != nil && req.OrderByTestSetNameDesc != nil {
		return nil, apierrors.ErrPagingTestCases.InvalidParameter("order by testSetName ASC or DESC?")
	}
	//?????????????????????,??????????????????test_id?????????test_set??????????????????????????????????????????test_set??????????????????
	if req.OrderByTestSetIDAsc == nil && req.OrderByTestSetIDDesc == nil &&
		req.OrderByTestSetNameAsc == nil && req.OrderByTestSetNameDesc == nil {
		// default order by `test_set_id` ASC
		req.OrderByTestSetIDAsc = &[]bool{true}[0]
		req.OrderFields = append(req.OrderFields, fieldTestSetID)
	}

	if req.OrderByPriorityAsc == nil && req.OrderByPriorityDesc == nil &&
		req.OrderByUpdaterIDAsc == nil && req.OrderByUpdaterIDDesc == nil &&
		req.OrderByUpdatedAtAsc == nil && req.OrderByUpdatedAtDesc == nil &&
		req.OrderByIDAsc == nil && req.OrderByIDDesc == nil {
		// default order by `id` ASC
		req.OrderByIDAsc = &[]bool{true}[0]
		req.OrderFields = append(req.OrderFields, fieldID)
	}

	if req.PageNo == 0 {
		req.PageNo = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// ???????????? testSet ??????????????? testSet
	allTestSetIDs, allTestSets, err := svc.db.ListTestSetsRecursive(apistructs.TestSetListRequest{
		Recycled:      req.Recycled,
		ParentID:      &req.TestSetID,
		ProjectID:     &req.ProjectID,
		TestSetIDs:    nil,
		NoSubTestSets: req.NoSubTestSet,
	})
	if err != nil {
		return nil, apierrors.ErrPagingTestCases.InternalError(
			fmt.Errorf("failed to get all children testSet, testSetID: %d, projectID: %d, err: %v", req.TestSetID, req.ProjectID, err))
	}
	// ????????????????????????????????????
	if len(allTestSetIDs) == 0 {
		return &apistructs.TestCasePagingResponseData{Total: 0, TestSets: nil, UserIDs: req.UpdaterIDs}, nil
	}

	sql := svc.db.DB

	if req.TestSetCaseMap != nil && len(req.TestSetCaseMap) > 0 {
		// ??????????????????????????????TestCaseIDs??????????????????
		caseIDs := []uint64{}
		for _, testSetID := range allTestSetIDs {
			_, exist := req.TestSetCaseMap[testSetID]
			if exist {
				caseIDs = append(caseIDs, req.TestSetCaseMap[testSetID]...)
			}
		}
		// ?????? ID ?????????????????????????????????
		// ?????????orm ??????????????????????????????????????? caseIDs ????????????
		if len(caseIDs) == 0 {
			return &apistructs.TestCasePagingResponseData{Total: 0, TestSets: nil}, nil
		}
		req.TestCaseIDs = caseIDs
	} else {
		// ????????????????????????????????????????????????
		sql = sql.Where("`test_set_id` IN (?)", allTestSetIDs)
	}

	// ???????????????????????????????????????????????????
	req.TestCaseIDs = strutil.DedupUint64Slice(req.TestCaseIDs, true)
	if len(req.TestCaseIDs) > 0 {
		sql = sql.Where("`id` IN (?)", req.TestCaseIDs)
	}

	// ??????????????????????????????????????????
	if len(req.NotInTestPlanIDs) > 0 {
		notInRels, err := svc.db.ListTestPlanCaseRels(apistructs.TestPlanCaseRelListRequest{
			TestPlanIDs:  req.NotInTestPlanIDs,
			IDOnly:       false,
			IdentityInfo: req.IdentityInfo,
		})
		if err != nil {
			return nil, apierrors.ErrPagingTestPlanCaseRels.InternalError(err)
		}
		for _, NotInRel := range notInRels {
			req.NotInTestCaseIDs = append(req.NotInTestCaseIDs, NotInRel.TestCaseID)
		}
	}
	req.NotInTestCaseIDs = strutil.DedupUint64Slice(req.NotInTestCaseIDs, true)
	if len(req.NotInTestCaseIDs) > 0 {
		sql = sql.Where("`id` NOT IN (?)", req.NotInTestCaseIDs)
	}

	// ????????????
	sql = sql.Where("`recycled` = ?", req.Recycled)
	// ?????? ??????
	sql = sql.Where("`project_id` = ?", req.ProjectID)
	// ?????? ??????
	if req.Query != "" {
		sql = sql.Where("`name` LIKE ?", strutil.Concat("%", req.Query, "%"))
	}
	// ????????? ????????????
	if len(req.Priorities) > 0 {
		sql = sql.Where("`priority` IN (?)", req.Priorities)
	}
	// ????????? ????????????
	if len(req.UpdaterIDs) > 0 {
		sql = sql.Where("`updater_id` IN (?)", req.UpdaterIDs)
	}
	// ???????????????????????? ??????????????????
	if req.TimestampSecUpdatedAtBegin != nil {
		t := time.Unix(int64(*req.TimestampSecUpdatedAtBegin), 0)
		req.UpdatedAtBeginInclude = &t
	}
	if req.UpdatedAtBeginInclude != nil {
		sql = sql.Where("`updated_at` >= ?", req.UpdatedAtBeginInclude)
	}
	// ???????????????????????? ??????????????????
	if req.TimestampSecUpdatedAtEnd != nil {
		t := time.Unix(int64(*req.TimestampSecUpdatedAtEnd), 0)
		req.UpdatedAtEndInclude = &t
	}
	if req.UpdatedAtEndInclude != nil {
		sql = sql.Where("`updated_at` <= ?", req.UpdatedAtEndInclude)
	}
	if len(req.OrderFields) == 0 {
		req.OrderFields = []string{fieldPriority, fieldID, fieldTestSetID, fieldUpdaterID, fieldUpdatedAt}
	}
	orderConds := make(map[string]string) // key: field, value: order condition
	// order by priority
	if req.OrderByPriorityAsc != nil {
		orderConds[fieldPriority] = "`priority` ASC"
	}
	if req.OrderByPriorityDesc != nil {
		orderConds[fieldPriority] = "`priority` DESC"
	}
	if req.OrderByUpdaterIDAsc != nil {
		orderConds[fieldUpdaterID] = "`updater_id` ASC"
	}
	if req.OrderByUpdaterIDDesc != nil {
		orderConds[fieldUpdaterID] = "`updater_id` DESC"
	}
	if req.OrderByUpdatedAtAsc != nil {
		orderConds[fieldUpdatedAt] = "`updated_at` ASC"
	}
	if req.OrderByUpdatedAtDesc != nil {
		orderConds[fieldUpdatedAt] = "`updated_at` DESC"
	}
	if req.OrderByIDAsc != nil {
		orderConds[fieldID] = "`id` ASC"
	}
	if req.OrderByIDDesc != nil {
		orderConds[fieldID] = "`id` DESC"
	}
	// ????????????????????????????????????????????????????????????????????????????????????????????? `test_set_id` ???????????????
	// ??? ORDER BY FIELD(`test_set_id`, id1, id2, id3...)
	if req.OrderByTestSetIDAsc != nil || req.OrderByTestSetIDDesc != nil {
		order := "ASC"
		if req.OrderByTestSetIDDesc != nil {
			order = "DESC"
		}
		orderConds[fieldTestSetID] = fmt.Sprintf("`test_set_id` %s", order)
	}
	if req.OrderByTestSetNameAsc != nil || req.OrderByTestSetNameDesc != nil {
		order := "ASC"
		if req.OrderByTestSetNameDesc != nil {
			order = "DESC"
		}
		sortedTestSetIDs := getAlphabetSortedTestSetIDs(allTestSets, order)
		sortedTestSetIDStrs := make([]string, 0, len(sortedTestSetIDs))
		for _, id := range sortedTestSetIDs {
			sortedTestSetIDStrs = append(sortedTestSetIDStrs, fmt.Sprintf("%d", id))
		}
		if len(sortedTestSetIDStrs) > 0 {
			orderConds[fieldTestSetID] = fmt.Sprintf("FIELD (`test_set_id`, %s)", strutil.Join(sortedTestSetIDStrs, ","))
		}
	}
	req.OrderFields = strutil.DedupSlice(req.OrderFields)
	for _, field := range req.OrderFields {
		if orderCond, ok := orderConds[field]; ok {
			sql = sql.Order(orderCond)
		}
	}

	// ??????
	var (
		testCases []dao.TestCase
		total     uint64
	)

	// offset, limit
	if req.PageNo == -1 && req.PageSize == -1 {
		// fetch all records, get total from results
		sql = sql.Find(&testCases)
	} else {
		// fetch requested page number
		offset := (req.PageNo - 1) * req.PageSize
		limit := req.PageSize
		sql = sql.Offset(offset).Limit(limit).Find(&testCases)
		// reset offset & limit before count
		sql = sql.Offset(0).Limit(-1).Count(&total)
	}

	// ?????? sql
	if err := sql.Error; err != nil {
		return nil, apierrors.ErrPagingTestCases.InternalError(err)
	}

	if req.PageNo == -1 && req.PageSize == -1 {
		total = uint64(len(testCases))
	}

	//??????????????????????????????????????????,???????????????
	if len(testCases) == 0 {
		return &apistructs.TestCasePagingResponseData{Total: total, TestSets: nil, UserIDs: req.UpdaterIDs}, nil
	}

	// ??? ?????????????????? ????????? ?????????(??????????????????)??????
	mapOfTestSetIDAndDir := make(map[uint64]string)
	for _, ts := range allTestSets {
		mapOfTestSetIDAndDir[ts.ID] = ts.Directory
	}
	resultTestSetMap := make(map[uint64]apistructs.TestSetWithCases)
	var testSetIDOrdered []uint64

	// ??? ???????????? ??????????????? ?????????
	// batchConvert testCases
	convertedTCs, err := svc.batchConvertTestCases(req.ProjectID, testCases)
	if err != nil {
		return nil, err
	}
	// map: ts.ID -> TestSetWithCases ([]tc)
	for i, tc := range testCases {
		// testSetID ??????
		if _, ok := resultTestSetMap[tc.TestSetID]; !ok {
			testSetIDOrdered = append(testSetIDOrdered, tc.TestSetID)
		}
		// testSetWithCase ????????????
		tmp := resultTestSetMap[tc.TestSetID]
		tmp.Directory = mapOfTestSetIDAndDir[tc.TestSetID]
		tmp.TestSetID = tc.TestSetID
		tmp.TestCases = append(tmp.TestCases, *convertedTCs[i])
		resultTestSetMap[tc.TestSetID] = tmp
	}
	resultTestSets := make([]apistructs.TestSetWithCases, 0)
	for _, tsID := range testSetIDOrdered {
		if ts, ok := resultTestSetMap[tsID]; ok {
			resultTestSets = append(resultTestSets, ts)
		}
	}

	// ?????????????????? ID
	var allUserIDs []string
	for _, ts := range resultTestSets {
		for _, tc := range ts.TestCases {
			allUserIDs = append(allUserIDs, tc.CreatorID, tc.UpdaterID)
		}
	}
	allUserIDs = strutil.DedupSlice(allUserIDs, true)

	// ????????????
	result := apistructs.TestCasePagingResponseData{
		Total:    total,
		TestSets: resultTestSets,
		UserIDs:  allUserIDs,
	}

	return &result, nil
}

// getAlphabetSortedTestSetIDs ???????????????????????????????????? testSetID ??????
// 1: /z       2: /ab
// 2: /ab  =>  3: /c
// 3: /c       1: /z
func getAlphabetSortedTestSetIDs(testSets []dao.TestSet, order string) []uint64 {
	m := make(map[string]uint64) // key: dir, value: testSetID
	var allTestSetDirs []string
	for _, ts := range testSets {
		m[ts.Directory] = ts.ID
		allTestSetDirs = append(allTestSetDirs, ts.Directory)
	}
	allTestSetDirs = strutil.DedupSlice(allTestSetDirs, true)
	sort.Strings(allTestSetDirs)

	var result []uint64
	for _, dir := range allTestSetDirs {
		result = append(result, m[dir])
	}
	return result
}
