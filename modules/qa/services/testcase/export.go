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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/modules/qa/services/i18n"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/xmind"
)

const maxAllowedNumberForTestCaseNumberExport = 2000

func (svc *Service) Export(w io.Writer, req apistructs.TestCaseExportRequest) error {
	// 参数校验
	if !req.FileType.Valid() {
		return apierrors.ErrExportTestCases.InvalidParameter("fileType")
	}
	beginPaging := time.Now()
	// 根据分页查询条件，获取总数，进行优化
	req.PageNo = -1
	req.PageSize = -1
	totalResult, err := svc.PagingTestCases(req.TestCasePagingRequest)
	if err != nil {
		return err
	}
	endPaging := time.Now()
	logrus.Debugf("export paging testcases cost: %fs", endPaging.Sub(beginPaging).Seconds())

	// limit
	if totalResult.Total > maxAllowedNumberForTestCaseNumberExport && req.FileType == apistructs.TestCaseFileTypeExcel {
		return apierrors.ErrExportTestCases.InvalidParameter(
			fmt.Sprintf("to many testcases: %d, max allowed number for export excel is: %d, please use xmind", totalResult.Total, maxAllowedNumberForTestCaseNumberExport))
	}

	// 结果处理
	var testCases []apistructs.TestCaseWithSimpleSetInfo
	for _, ts := range totalResult.TestSets {
		for _, tc := range ts.TestCases {
			testCases = append(testCases, apistructs.TestCaseWithSimpleSetInfo{TestCase: tc, Directory: ts.Directory})
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
