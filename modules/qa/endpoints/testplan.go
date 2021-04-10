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
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	urlPathTestPlanID        = "testPlanID"
	urlPathTestPlanCaseRelID = "relationID"
)

// CreateTestPlan 创建测试计划
func (e *Endpoints) CreateTestPlan(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateTestPlan.NotLogin().ToResp(), nil
	}

	var req apistructs.TestPlanCreateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrCreateTestPlan.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateTestPlan.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	testPlanID, err := e.testPlan.Create(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testPlanID)
}

// UpdateTestPlan 更新测试计划
func (e *Endpoints) UpdateTestPlan(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateTestPlan.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateTestPlan.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanUpdateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateTestPlan.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateTestPlan.InvalidParameter(err).ToResp(), nil
	}
	req.TestPlanID = testPlanID
	req.IdentityInfo = identityInfo

	// 查询测试计划用于鉴权
	tp, err := e.testPlan.Get(req.TestPlanID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 鉴权
	if !req.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  tp.ProjectID,
			Resource: apistructs.TestPlanResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return apierrors.ErrCheckPermission.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrDeleteTestPlan.AccessDenied().ToResp(), nil
		}
	}

	if err := e.testPlan.Update(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// GetTestPlan 测试计划详情
func (e *Endpoints) GetTestPlan(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetTestPlan.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrGetTestPlan.InvalidParameter(err).ToResp(), nil
	}

	testPlan, err := e.testPlan.Get(testPlanID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	userIDs := strutil.DedupSlice(append([]string{testPlan.OwnerID, testPlan.CreatorID, testPlan.UpdaterID}, testPlan.PartnerIDs...))

	return httpserver.OkResp(testPlan, userIDs)
}

// DeleteTestPlan 删除测试计划
func (e *Endpoints) DeleteTestPlan(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteTestPlan.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteTestPlan.InvalidParameter(err).ToResp(), nil
	}

	if err := e.testPlan.Delete(identityInfo, testPlanID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testPlanID)
}

// PagingTestPlans 测试计划分页
func (e *Endpoints) PagingTestPlans(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingTestPlans.NotLogin().ToResp(), nil
	}
	var req apistructs.TestPlanPagingRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingTestPlans.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	if req.ProjectID == 0 {
		return apierrors.ErrPagingTestPlans.MissingParameter("projectID").ToResp(), nil
	}

	if !req.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectID,
			Resource: apistructs.TestPlanResource,
			Action:   apistructs.ListAction,
		})
		if err != nil {
			return nil, apierrors.ErrPagingTestPlans.InternalError(err)
		}
		if !access.Access {
			return nil, apierrors.ErrPagingTestPlans.AccessDenied()
		}
	}

	result, err := e.testPlan.Paging(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(result, result.UserIDs)
}

// ListTestPlanUsecase 测试计划内测试用例过滤
func (e *Endpoints) PagingTestPlanCaseRelations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingTestPlanCaseRels.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrPagingTestPlanCaseRels.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanCaseRelPagingRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingTestPlans.InvalidParameter(err).ToResp(), nil
	}
	req.TestPlanID = testPlanID
	req.IdentityInfo = identityInfo

	// 查询测试计划
	tp, err := e.testPlan.Get(req.TestPlanID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if !req.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  tp.ProjectID,
			Resource: apistructs.TestPlanResource,
			Action:   apistructs.ListAction,
		})
		if err != nil {
			return nil, err
		}
		if !access.Access {
			return nil, apierrors.ErrPagingTestPlanCaseRels.AccessDenied()
		}
	}

	result, err := e.testPlan.PagingTestPlanCaseRels(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(result, result.UserIDs)
}

// InternalListTestPlanCaseRels 仅供内部使用
func (e *Endpoints) InternalListTestPlanCaseRels(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingTestPlanCaseRels.NotLogin().ToResp(), nil
	}

	var req apistructs.TestPlanCaseRelListRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingTestPlanCaseRels.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	rels, err := e.testPlan.ListTestPlanCaseRels(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(rels)
}

// ExecuteTestPlanAPITest 执行测试计划接口测试
func (e *Endpoints) ExecuteTestPlanAPITest(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrTestPlanExecuteAPITest.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrTestPlanExecuteAPITest.InvalidParameter(err).ToResp(), nil
	}

	if r.ContentLength == 0 {
		return apierrors.ErrTestPlanExecuteAPITest.MissingParameter("request body").ToResp(), nil
	}
	var req apistructs.TestPlanAPITestExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrTestPlanExecuteAPITest.InvalidParameter(err).ToResp(), nil
	}
	req.TestPlanID = testPlanID
	req.IdentityInfo = identityInfo

	tp, err := e.testPlan.Get(req.TestPlanID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if !req.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  tp.ProjectID,
			Resource: apistructs.TestPlanResource,
			Action:   apistructs.OperateAction,
		})
		if err != nil {
			return errorresp.ErrResp(err)
		}
		if !access.Access {
			return apierrors.ErrTestPlanExecuteAPITest.AccessDenied().ToResp(), nil
		}
	}

	triggeredPipelineID, err := e.testPlan.ExecuteAPITest(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(triggeredPipelineID)
}

// ListTestPlanTestSets 获取测试计划下的测试集列表
func (e *Endpoints) ListTestPlanTestSets(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListTestPlanTestSets.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrListTestPlanTestSets.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanTestSetsListRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListTestPlanTestSets.InvalidParameter(err).ToResp(), nil
	}
	req.TestPlanID = testPlanID
	req.IdentityInfo = identityInfo

	testSets, err := e.testPlan.ListTestSet(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testSets)
}

// ExportTestCases 导出测试计划下的测试用例
func (e *Endpoints) ExportTestPlanCaseRels(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrExportTestPlanCaseRels.NotLogin()
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrExportTestPlanCaseRels.InvalidParameter(err)
	}

	var req apistructs.TestPlanCaseRelExportRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrExportTestPlanCaseRels.InvalidParameter(err)
	}
	req.TestPlanID = testPlanID
	req.IdentityInfo = identityInfo

	l := e.bdl.GetLocaleByRequest(r)
	req.Locale = l.Name()

	// TODO:鉴权

	err = e.testPlan.Export(w, req)
	if err != nil {
		return apierrors.ErrExportTestCases.InternalError(err)
	}

	return nil
}

// GenerateTestPlanReport 生成测试计划报告
func (e *Endpoints) GenerateTestPlanReport(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGenerateTestPlanReport.NotLogin().ToResp(), nil
	}
	_ = identityInfo

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrGenerateTestPlanReport.InvalidParameter(err).ToResp(), nil
	}

	report, err := e.testPlan.GenerateReport(testPlanID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(report, report.UserIDs)
}
