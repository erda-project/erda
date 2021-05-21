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
	"io"
	"sync"

	"github.com/mohae/deepcopy"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/modules/qa/services/i18n"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/xmind"
)

func (svc *Service) Export(w io.Writer, req apistructs.TestCaseExportRequest) error {
	// 参数校验
	if !req.FileType.Valid() {
		return apierrors.ErrExportTestCases.InvalidParameter("fileType")
	}
	// 根据分页查询条件，获取总数，进行优化
	req.PageNo = 1
	req.PageSize = 1
	totalResult, err := svc.PagingTestCases(req.TestCasePagingRequest)
	if err != nil {
		return err
	}
	total := totalResult.Total

	// 以 size=200 并行获取 testCase，加快速度
	pageSize := uint64(200)
	count := int(total)/int(pageSize) + 1
	var wg sync.WaitGroup
	var errs []string
	testSetMap := make(map[int][]apistructs.TestSetWithCases)
	for i := 0; i < count; i++ {
		c := deepcopy.Copy(req.TestCasePagingRequest)
		copiedReq, ok := c.(apistructs.TestCasePagingRequest)
		if !ok {
			panic("should not be here")
		}
		copiedReq.PageNo = uint64(i + 1)
		copiedReq.PageSize = pageSize

		wg.Add(1)
		go func(order int, req apistructs.TestCasePagingRequest, testSetMap map[int][]apistructs.TestSetWithCases) {
			defer wg.Done()
			pagingResult, err := svc.PagingTestCases(req)
			if err != nil {
				errs = append(errs, err.Error())
				return
			}
			testSetMap[order] = pagingResult.TestSets
		}(i, copiedReq, testSetMap)
	}
	wg.Wait()

	// 错误处理
	if len(errs) > 0 {
		return fmt.Errorf(strutil.Join(errs, ",", true))
	}

	// 结果处理
	var testCases []apistructs.TestCaseWithSimpleSetInfo
	testSets := make([][]apistructs.TestSetWithCases, 0, count)
	for i := 0; i < count; i++ {
		testSets = append(testSets, testSetMap[i])
	}
	for _, testSets := range testSets {
		for _, ts := range testSets {
			for _, tc := range ts.TestCases {
				testCases = append(testCases, apistructs.TestCaseWithSimpleSetInfo{TestCase: tc, Directory: ts.Directory})
			}
		}
	}

	l := svc.bdl.GetLocale(req.Locale)

	sheetName := l.Get(i18n.I18nKeyTestCaseSheetName)

	// export
	switch req.FileType {
	case apistructs.TestCaseFileTypeExcel:
		excelLines, err := svc.convert2Excel(testCases, req.Locale)
		if err != nil {
			return err
		}
		return excel.ExportExcelByCell(w, excelLines, sheetName)
	case apistructs.TestCaseFileTypeXmind:
		xmindContent, err := svc.convert2XMind(testCases, req.Locale)
		if err != nil {
			return err
		}
		return xmind.Export(w, xmindContent, sheetName)
	default:
		return fmt.Errorf("export type %s is not supported yet", req.FileType)
	}
}
