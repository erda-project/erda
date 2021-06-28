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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/i18n"
	"github.com/erda-project/erda/pkg/excel"
	"github.com/erda-project/erda/pkg/xmind"
)

const (
	maxAllowedNumberForTestCaseNumberExport = 2000
	defaultResource                         = "testcases"
)

func (svc *Service) Export(req apistructs.TestCaseExportRequest) (uint64, error) {
	// 参数校验
	if !req.FileType.Valid() {
		return 0, apierrors.ErrExportTestCases.InvalidParameter("fileType")
	}

	beginPaging := time.Now()
	// 根据分页查询条件，获取总数，进行优化
	req.PageNo = -1
	req.PageSize = -1
	totalResult, err := svc.PagingTestCases(req.TestCasePagingRequest)
	if err != nil {
		return 0, err
	}
	endPaging := time.Now()
	logrus.Debugf("export paging testcases cost: %fs", endPaging.Sub(beginPaging).Seconds())

	// limit
	if totalResult.Total > maxAllowedNumberForTestCaseNumberExport && req.FileType == apistructs.TestCaseFileTypeExcel {
		return 0, apierrors.ErrExportTestCases.InvalidParameter(
			fmt.Sprintf("to many testcases: %d, max allowed number for export excel is: %d, please use xmind", totalResult.Total, maxAllowedNumberForTestCaseNumberExport))
	}
	l := svc.bdl.GetLocale(req.Locale)
	sheetName := l.Get(i18n.I18nKeyTestCaseSheetName, defaultResource)
	if req.FileType == apistructs.TestCaseFileTypeExcel {
		sheetName += ".xlsx"
	} else {
		sheetName += ".xmind"
	}
	fileReq := apistructs.TestFileRecordRequest{
		FileName:     sheetName,
		Description:  fmt.Sprintf("ProjectID: %v, TestsetID: %v", req.ProjectID, req.TestSetID),
		ProjectID:    req.ProjectID,
		TestSetID:    req.TestSetID,
		Type:         apistructs.FileActionTypeExport,
		State:        apistructs.FileRecordStatePending,
		IdentityInfo: req.IdentityInfo,
		Extra: apistructs.TestFileExtra{
			ManualTestFileExtraInfo: &apistructs.ManualTestFileExtraInfo{
				ExportRequest: &req,
			},
		},
	}
	id, err := svc.CreateFileRecord(fileReq)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (svc *Service) ExportFile(record *dao.TestFileRecord) {
	req := record.Extra.ManualTestFileExtraInfo.ExportRequest
	req.PageNo = -1
	req.PageSize = -1
	totalResult, err := svc.PagingTestCases(req.TestCasePagingRequest)
	if err != nil {
		logrus.Error(err)
		return
	}

	var testCases []apistructs.TestCaseWithSimpleSetInfo
	for _, ts := range totalResult.TestSets {
		for _, tc := range ts.TestCases {
			testCases = append(testCases, apistructs.TestCaseWithSimpleSetInfo{TestCase: tc, Directory: ts.Directory})
		}
	}

	id := record.ID
	sheetName := record.FileName
	if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateProcessing}); err != nil {
		logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
		return
	}

	if req.FileType == apistructs.TestCaseFileTypeExcel {
		excelLines, err := svc.convert2Excel(testCases, req.Locale)
		if err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			}
			return
		}

		buff, err := excel.WriteExcelBuffer(excelLines, sheetName)
		if err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			}
			return
		}

		uuid, err := svc.Upload(buff, sheetName)
		if err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			}
			return
		}

		if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateSuccess, ApiFileUUID: uuid}); err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
		}
	} else {
		xmindContent, err := svc.convert2XMind(testCases, req.Locale)
		if err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			}
			return
		}

		f, err := ioutil.TempFile("", "export.*.xmind")
		if err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			return
		}

		defer func() {
			if err := os.Remove(f.Name()); err != nil {
				logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			}
		}()

		defer f.Close()

		err = xmind.Export(f, xmindContent, sheetName)
		if err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			}
			return
		}

		req := apistructs.FileUploadRequest{
			FileNameWithExt: sheetName,
			FileReader:      f,
			From:            defaultResource,
			IsPublic:        true,
			ExpiredAt:       nil,
		}
		file, err := svc.bdl.UploadFile(req)
		if err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
			}
			return
		}

		if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateSuccess, ApiFileUUID: file.UUID}); err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
		}
	}
}

func (svc *Service) Upload(buff *bytes.Buffer, fileName string) (string, error) {
	req := apistructs.FileUploadRequest{
		FileNameWithExt: fileName,
		ByteSize:        int64(buff.Len()),
		FileReader:      ioutil.NopCloser(buff),
		From:            defaultResource,
		IsPublic:        true,
		Encrypt:         false,
		ExpiredAt:       nil,
	}
	res, err := svc.bdl.UploadFile(req)
	if err != nil {
		return "", err
	}
	return res.UUID, nil
}
