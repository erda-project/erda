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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
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

	// 获取测试用例数据
	f, _, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	switch req.FileType {
	case apistructs.TestCaseFileTypeExcel:
		excelTcs, err := svc.decodeFromExcelFile(f)
		if err != nil {
			return nil, err
		}
		return svc.storeExcel2DB(req, ts, excelTcs)
	case apistructs.TestCaseFileTypeXmind:
		xmindTcs, err := svc.decodeFromXMindFile(f)
		if err != nil {
			return nil, err
		}
		return svc.storeXmind2DB(req, ts, xmindTcs)
	default:
		return nil, fmt.Errorf("import from %s is not supported yet", req.FileType)
	}
}
