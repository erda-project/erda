// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testcase

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/internal/apps/dop/services/i18n"
	"github.com/erda-project/erda/internal/core/file/filetypes"
	"github.com/erda-project/erda/pkg/database/dbengine"
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
		Type:         apistructs.FileActionTypeExport,
		State:        apistructs.FileRecordStatePending,
		IdentityInfo: req.IdentityInfo,
		Extra: apistructs.TestFileExtra{
			ManualTestFileExtraInfo: &apistructs.ManualTestFileExtraInfo{
				TestSetID:     req.TestSetID,
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
	id := record.ID
	if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateProcessing}); err != nil {
		logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
		return
	}
	fileUUID, err := svc.ExportTestCases(req, record.FileName, true)
	if err != nil {
		logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
		if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateFail, ErrorInfo: err}); err != nil {
			logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
		}
		return
	}
	if err := svc.UpdateFileRecord(apistructs.TestFileRecordRequest{ID: id, State: apistructs.FileRecordStateSuccess, ApiFileUUID: fileUUID}); err != nil {
		logrus.Error(apierrors.ErrImportTestCases.InternalError(err))
	}
}

// needQuery 参数表示测试用例是否从数据库获取
// 如果是从测试用例页面导出，这里需要查表， needQuery 设置为 true
// 如果是从 AI 生成测试用例页面导出，则无需查表，因为生成的结果并没有存储到数据库， needQuery 设置为 false
func (svc *Service) ExportTestCases(req *apistructs.TestCaseExportRequest, sheetName string, needQuery bool) (string, error) {
	req.PageNo = -1
	req.PageSize = -1
	var totalResult *apistructs.TestCasePagingResponseData
	var err error
	if needQuery {
		totalResult, err = svc.PagingTestCases(req.TestCasePagingRequest)
		if err != nil {
			return "", err
		}
	} else {
		// AI 生成测试用例页面导出，因此无需查表
		totalResult, err = svc.generateTestCasePagingResponseData(req)
		if err != nil {
			return "", err
		}
	}

	var testCases []apistructs.TestCaseWithSimpleSetInfo
	for _, ts := range totalResult.TestSets {
		for _, tc := range ts.TestCases {
			testCases = append(testCases, apistructs.TestCaseWithSimpleSetInfo{TestCase: tc, Directory: ts.Directory})
		}
	}

	f, err := os.CreateTemp("", "export.*")
	if err != nil {
		return "", apierrors.ErrExportTestCases.InternalError(err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			logrus.Error(apierrors.ErrExportTestCases.InternalError(err))
		}
	}()

	defer f.Close()

	if req.FileType == apistructs.TestCaseFileTypeExcel {
		excelLines, err := svc.convert2Excel(testCases, req.Locale)
		if err != nil {
			return "", apierrors.ErrExportTestCases.InternalError(err)
		}

		if err := excel.ExportExcelByCell(f, excelLines, sheetName); err != nil {
			return "", apierrors.ErrExportTestCases.InternalError(err)
		}
	} else {
		xmindContent, err := svc.convert2XMind(testCases, req.Locale)
		if err != nil {
			return "", apierrors.ErrExportTestCases.InternalError(err)
		}

		if err := xmind.Export(f, xmindContent, sheetName); err != nil {
			return "", apierrors.ErrExportTestCases.InternalError(err)
		}
	}

	//Set offset for next read
	f.Seek(0, 0)

	uploadReq := filetypes.FileUploadRequest{
		FileNameWithExt: sheetName,
		FileReader:      f,
		From:            defaultResource,
		IsPublic:        true,
		ExpiredAt:       nil,
	}
	file, err := svc.bdl.UploadFile(uploadReq)
	if err != nil {
		return "", apierrors.ErrExportTestCases.InternalError(err)
	}
	return file.UUID, nil
}

func (svc *Service) ExportAIGenerated(req apistructs.TestCaseExportRequest) (uint64, error) {
	// 参数校验
	if !req.FileType.Valid() {
		return 0, apierrors.ErrExportTestCases.InvalidParameter("fileType")
	}

	if len(req.TestSetCasesMetas) == 0 {
		return 0, apierrors.ErrExportTestCases.InvalidParameter("testSetCasesMetas")
	}
	totalTestcase := 0
	for idx, tcm := range req.TestSetCasesMetas {
		if len(tcm.Reqs) == 0 {
			return 0, apierrors.ErrExportTestCases.InvalidParameter(fmt.Sprintf("testSetCasesMetas[%d].testCaseCreateReqs", idx))
		}
		for idy, tcmreq := range tcm.Reqs {
			if len(tcmreq.StepAndResults) == 0 {
				return 0, apierrors.ErrExportTestCases.InvalidParameter(fmt.Sprintf("testSetCasesMetas[%d].testCaseCreateReqs[%d].stepAndResults", idx, idy))
			}
		}
		totalTestcase = totalTestcase + len(tcm.Reqs)
	}

	// limit
	if totalTestcase > maxAllowedNumberForTestCaseNumberExport && req.FileType == apistructs.TestCaseFileTypeExcel {
		return 0, apierrors.ErrExportTestCases.InvalidParameter(
			fmt.Sprintf("to many testcases: %d, max allowed number for export excel is: %d, please use xmind", totalTestcase, maxAllowedNumberForTestCaseNumberExport))
	}

	reqNoTestCaseMetas := apistructs.TestCaseExportRequest{
		TestCasePagingRequest: req.TestCasePagingRequest,
		FileType:              req.FileType,
		Locale:                req.Locale,
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
		Type:         apistructs.FileActionTypeExport,
		State:        apistructs.FileRecordStatePending,
		IdentityInfo: req.IdentityInfo,
		Extra: apistructs.TestFileExtra{
			ManualTestFileExtraInfo: &apistructs.ManualTestFileExtraInfo{
				TestSetID:     req.TestSetID,
				ExportRequest: &reqNoTestCaseMetas,
			},
		},
	}

	id, err := svc.CreateFileRecord(fileReq)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (svc *Service) generateTestCasePagingResponseData(req *apistructs.TestCaseExportRequest) (*apistructs.TestCasePagingResponseData, error) {

	totalTestcase := 0
	for _, tcm := range req.TestSetCasesMetas {
		totalTestcase = totalTestcase + len(tcm.Reqs)
	}

	totalResult := &apistructs.TestCasePagingResponseData{
		Total:    uint64(totalTestcase),
		TestSets: nil,
		UserIDs:  nil,
	}

	//l2IdToL3IdsToTcs := make(map[uint64][]map[uint64][]apistructs.TestCase)

	//setIDToCasesx := make(map[uint64][]apistructs.TestCase)
	testcases := make([]apistructs.TestCase, 0)
	testSetsMap := make(map[uint64]dao.TestSet)
	for _, tsc := range req.TestSetCasesMetas {
		for _, tcm := range tsc.Reqs {
			/*
				if _, ok := setIDToCases[tcm.TestSetID]; !ok {
					setIDToCases[tcm.TestSetID] = make([]apistructs.TestCase, 0)
				}
			*/

			// 父测试集
			testSetsMap[tcm.ParentTestSetID] = dao.TestSet{
				BaseModel: dbengine.BaseModel{
					ID: tcm.TestSetID,
				},
				Name:      tcm.ParentTestSetDir,
				ParentID:  0,
				Recycled:  false,
				ProjectID: tcm.ProjectID,
				Directory: tcm.ParentTestSetDir,
				OrderNum:  0,
			}
			//当前测试集
			testSetsMap[tcm.TestSetID] = dao.TestSet{
				BaseModel: dbengine.BaseModel{
					ID: tcm.TestSetID,
				},
				Name:      tcm.TestSetName,
				ParentID:  tcm.ParentTestSetID,
				Recycled:  false,
				ProjectID: tcm.ProjectID,
				Directory: tcm.TestSetDir,
				OrderNum:  1,
			}
			recycled := false
			tc := apistructs.TestCase{
				ID:             tcm.TestCaseID, // 由于是 AI 生成的，还没有落到数据库，因此实际没有 ID
				Name:           tcm.Name,
				Priority:       tcm.Priority,
				PreCondition:   tcm.PreCondition,
				Desc:           tcm.Desc,
				Recycled:       &recycled,
				TestSetID:      tcm.TestSetID,
				ProjectID:      tcm.ProjectID,
				CreatorID:      tcm.UserID,
				UpdaterID:      tcm.UserID,
				StepAndResults: tcm.StepAndResults,
			}
			//setIDToCases[tcm.TestSetID] = append(setIDToCases[tcm.TestSetID], tc)
			testcases = append(testcases, tc)
		}
	}

	testSets := make([]dao.TestSet, 0)
	for _, v := range testSetsMap {
		testSets = append(testSets, v)
	}

	// 将 测试用例列表 转换为 测试集(包含测试用例)列表
	mapOfTestSetIDAndDir := make(map[uint64]string)
	for _, ts := range testSets {
		mapOfTestSetIDAndDir[ts.ID] = ts.Directory
	}
	resultTestSetMap := make(map[uint64]apistructs.TestSetWithCases)
	var testSetIDOrdered []uint64
	// map: ts.ID -> TestSetWithCases ([]tc)
	for i, tc := range testcases {
		// testSetID 排序
		if _, ok := resultTestSetMap[tc.TestSetID]; !ok {
			testSetIDOrdered = append(testSetIDOrdered, tc.TestSetID)
		}
		// testSetWithCase 内容填充
		tmp := resultTestSetMap[tc.TestSetID]
		tmp.Directory = mapOfTestSetIDAndDir[tc.TestSetID]
		tmp.TestSetID = tc.TestSetID
		tmp.TestCases = append(tmp.TestCases, testcases[i])
		resultTestSetMap[tc.TestSetID] = tmp
	}
	resultTestSets := make([]apistructs.TestSetWithCases, 0)
	for _, tsID := range testSetIDOrdered {
		if ts, ok := resultTestSetMap[tsID]; ok {
			resultTestSets = append(resultTestSets, ts)
		}
	}

	/*
	   	tss := make([]apistructs.TestSetWithCases, 0)


	   	for tsId, tcs := range setIDToCases {
	   		if _, ok := resultTestSetMap[tsId]; !ok {
	   			resultTestSetMap[tsId] = apistructs.TestSetWithCases{
	   				TestSetID: tsId,
	   				Recycled:  false,
	   				Directory: "",
	   				TestCases: nil,
	   			}
	   		}
	   		for _ ,tc := range tcs {

	   		}
	   		tss = append(tss, apistructs.TestSetWithCases{
	   			//TestSetID: tsId,
	   			TestSetID: tcs.,
	   			Recycled:  false,
	   			Directory: tcs[0].TestSetDir,
	   			TestCases: tcs,
	   		})
	   	}
	      totalResult.TestSets = tss
	*/

	totalResult.TestSets = resultTestSets
	// 因为是 AI 单次生成的用例，所以用户其实都只有一个，就是那个发起生成操作的用户
	totalResult.UserIDs = []string{req.UserID}

	return totalResult, nil
}
