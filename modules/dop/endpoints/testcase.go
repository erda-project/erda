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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateTestCase 创建测试用例
func (e *Endpoints) CreateTestCase(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateTestCase.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrCreateTestCase.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.TestCaseCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateTestCase.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// TODO:鉴权

	tcID, err := e.testcase.CreateTestCase(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(tcID)
}

func (e *Endpoints) BatchCreateTestCases(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrBatchCreateTestCases.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrBatchCreateTestCases.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.TestCaseBatchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrBatchCreateTestCases.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// TODO:鉴权

	testCaseIDs, err := e.testcase.BatchCreateTestCases(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testCaseIDs)
}

// UpdateTestCase 更新测试用例
func (e *Endpoints) UpdateTestCase(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateTestCase.NotLogin().ToResp(), nil
	}

	testCaseID, err := strconv.ParseUint(vars["testCaseID"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateTestCase.InvalidParameter("testCaseID").ToResp(), nil
	}

	// TODO:鉴权

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateTestCase.MissingParameter("request body").ToResp(), nil
	}
	var req apistructs.TestCaseUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateTestCase.InvalidParameter(err).ToResp(), nil
	}

	req.ID = testCaseID
	req.IdentityInfo = identityInfo
	if err := e.testcase.UpdateTestCase(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// GetTestCase 获取测试用例详情
func (e *Endpoints) GetTestCase(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	tcID, err := strconv.ParseUint(vars["testCaseID"], 10, 64)
	if err != nil {
		return apierrors.ErrGetTestCase.InvalidParameter("testCaseID").ToResp(), nil
	}

	// TODO: 操作鉴权

	tc, err := e.testcase.GetTestCase(tcID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(*tc, strutil.DedupSlice([]string{tc.CreatorID, tc.UpdaterID}, true))
}

// BatchUpdateTestCases 批量更新测试用例
func (e *Endpoints) BatchUpdateTestCases(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrBatchUpdateTestCases.NotLogin().ToResp(), nil
	}

	// TODO:鉴权

	// 校验 body 合法性
	var req apistructs.TestCaseBatchUpdateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrBatchUpdateTestCases.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrBatchUpdateTestCases.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	if err := e.testcase.BatchUpdateTestCases(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// BatchUpdateTestCases 批量复制测试用例
func (e *Endpoints) BatchCopyTestCases(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrBatchCopyTestCases.NotLogin().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrBatchCopyTestCases.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.TestCaseBatchCopyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrBatchCreateTestCases.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// TODO:鉴权

	copiedTestCaseIDs, err := e.testcase.BatchCopyTestCases(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(copiedTestCaseIDs)
}

func (e *Endpoints) BatchCleanTestCasesFromRecycleBin(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrBatchCleanTestCasesFromRecycleBin.NotLogin().ToResp(), nil
	}

	if r.ContentLength == 0 {
		return apierrors.ErrBatchCleanTestCasesFromRecycleBin.MissingParameter("request body").ToResp(), nil
	}
	var req apistructs.TestCaseBatchCleanFromRecycleBinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrBatchCleanTestCasesFromRecycleBin.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	if err := e.testcase.BatchCleanFromRecycleBin(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// PagingTestCases 获取测试用例列表
func (e *Endpoints) PagingTestCases(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.TestCasePagingRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingTestCases.InvalidParameter(err).ToResp(), nil
	}

	// TODO: 操作鉴权

	//判断UpdaterIDs在项目内是否有权限
	if len(req.UpdaterIDs) > 0 {
		members, _ := e.bdl.ListMembers(apistructs.MemberListRequest{
			ScopeType: apistructs.ProjectScope,
			ScopeID:   int64(req.ProjectID),
			PageNo:    1,
			PageSize:  300,
		})
		mapOfupdaterIDs := make(map[string]bool)
		for _, updater := range req.UpdaterIDs {
			mapOfupdaterIDs[updater] = false
		}
		for _, member := range members {
			if _, ok := mapOfupdaterIDs[member.UserID]; ok {
				mapOfupdaterIDs[member.UserID] = true
			}
		}
		for _, value := range mapOfupdaterIDs {
			if !value {
				return nil, apierrors.ErrPagingTestCases.AccessDenied()
			}
		}
	}

	pagingResult, err := e.testcase.PagingTestCases(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(pagingResult, pagingResult.UserIDs)
}

// ExportTestCases 导出测试用例
func (e *Endpoints) ExportTestCases(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrExportTestCases.NotLogin()
	}

	var req apistructs.TestCaseExportRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrExportTestCases.InvalidParameter(err)
	}
	req.IdentityInfo = identityInfo

	l := e.bdl.GetLocaleByRequest(r)
	req.Locale = l.Name()

	// TODO:鉴权

	err = e.testcase.Export(w, req)
	if err != nil {
		return apierrors.ErrExportTestCases.InternalError(err)
	}

	return nil
}

// ImportTestCases 导入测试用例
func (e *Endpoints) ImportTestCases(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrImportTestCases.NotLogin().ToResp(), nil
	}

	var req apistructs.TestCaseImportRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrImportTestCases.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	// TODO:鉴权

	importResult, err := e.testcase.Import(req, r)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(importResult)
}
