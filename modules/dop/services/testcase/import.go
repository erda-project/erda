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
	"net/http"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
)

// Import 导入测试用例
func (svc *Service) Import(req apistructs.TestCaseImportRequest, r *http.Request) (*apistructs.TestCaseImportResult, error) {
	// 参数校验
	if !req.FileType.Valid() {
		return nil, apierrors.ErrImportTestCases.InvalidParameter("fileType")
	}
	if req.ProjectID == 0 {
		return nil, apierrors.ErrImportTestCases.MissingParameter("projectID")
	}

	// fake ts
	ts := dao.FakeRootTestSet(req.ProjectID, false)
	if req.TestSetID != 0 {
		_ts, err := svc.db.GetTestSetByID(req.TestSetID)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return nil, apierrors.ErrImportTestCases.InvalidParameter(fmt.Errorf("testSet not found, id: %d", req.TestSetID))
			}
			return nil, apierrors.ErrImportTestCases.InternalError(err)
		}
		ts = *_ts
	}
	if ts.ProjectID != req.ProjectID {
		return nil, apierrors.ErrImportTestCases.InvalidParameter("projectID")
	}

	// get testsets data
	f, fileHeader, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	uploadReq := apistructs.FileUploadRequest{
		FileNameWithExt: fileHeader.Filename,
		FileReader:      f,
		From:            "testcase",
		IsPublic:        true,
		ExpiredAt:       nil,
	}
	file, err := svc.bdl.UploadFile(uploadReq)
	if err != nil {
		return nil, err
	}

	fileReq := apistructs.TestFileRecordRequest{
		FileName:     fileHeader.Filename,
		Description:  fmt.Sprintf("ProjectID: %v, TestsetID: %v", req.ProjectID, req.TestSetID),
		ProjectID:    req.ProjectID,
		TestSetID:    req.TestSetID,
		Type:         apistructs.FileActionTypeImport,
		ApiFileUUID:  file.UUID,
		State:        apistructs.FileRecordStatePending,
		IdentityInfo: req.IdentityInfo,
		Extra: apistructs.TestFileExtra{
			ManualTestFileExtraInfo: &apistructs.ManualTestFileExtraInfo{
				ImportRequest: &req,
			},
		},
	}
	id, err := svc.CreateFileRecord(fileReq)
	if err != nil {
		return nil, err
	}
	return &apistructs.TestCaseImportResult{Id: id}, nil
}

func (svc *Service) ImportFile(record *dao.TestFileRecord) {
	req := record.Extra.ManualTestFileExtraInfo.ImportRequest
	ts := dao.FakeRootTestSet(req.ProjectID, false)
	if req.TestSetID != 0 {
		_ts, err := svc.db.GetTestSetByID(req.TestSetID)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				logrus.Error(apierrors.ErrImportTestCases.InvalidParameter(fmt.Errorf("testSet not found, id: %d", req.TestSetID)))
				return
			}
			logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
			return
		}
		ts = *_ts
	}
	if ts.ProjectID != req.ProjectID {
		logrus.Error(apierrors.ErrImportTestCases.InvalidParameter("projectID"))
		return
	}

	id := record.ID
	if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateProcessing}); err != nil {
		logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
		return
	}

	f, err := svc.bdl.DownloadDiceFile(record.ApiFileUUID)
	if err != nil {
		logrus.Error(err)
		return
	}

	if req.FileType == apistructs.TestCaseFileTypeExcel {
		excelTcs, err := svc.decodeFromExcelFile(f)
		if err != nil {
			logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
			}
			return
		}
		if _, err := svc.storeExcel2DB(*req, ts, excelTcs); err != nil {
			logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
			return
		}
	} else {
		xmindTcs, err := svc.decodeFromXMindFile(f)
		if err != nil {
			logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
			if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail}); err != nil {
				logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
			}
			return
		}
		if _, err := svc.storeXmind2DB(*req, ts, xmindTcs); err != nil {
			logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
			return
		}
	}
	if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateSuccess}); err != nil {
		logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
	}
}
