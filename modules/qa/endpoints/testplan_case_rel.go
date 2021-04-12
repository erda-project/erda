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
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateTestPlanCaseRelations 创建测试计划用例关系
func (e *Endpoints) CreateTestPlanCaseRelations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateTestPlanCaseRel.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrCreateTestPlanCaseRel.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanCaseRelCreateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrCreateTestPlanCaseRel.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateTestPlanCaseRel.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	req.TestPlanID = testPlanID

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
			Resource: apistructs.TestPlanUsecaseRelResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return errorresp.ErrResp(err)
		}
		if !access.Access {
			return apierrors.ErrCreateTestPlanCaseRel.AccessDenied().ToResp(), nil
		}
	}

	result, err := e.testPlan.CreateCaseRelations(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) GetTestPlanCaseRel(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetTestPlanCaseRel.NotLogin().ToResp(), nil
	}
	_ = identityInfo

	relID, err := strconv.ParseUint(vars["relationID"], 10, 64)
	if err != nil {
		return apierrors.ErrGetTestPlanCaseRel.InvalidParameter("relationID").ToResp(), nil
	}

	rel, err := e.testPlan.GetRel(relID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(rel, strutil.DedupSlice([]string{rel.CreatorID, rel.UpdaterID, rel.ExecutorID}))
}

// BatchUpdateTestPlanCaseRelations 批量更新测试计划用例关系
func (e *Endpoints) BatchUpdateTestPlanCaseRelations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrBatchUpdateTestPlanCaseRels.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrBatchUpdateTestPlanCaseRels.InvalidParameter(err).ToResp(), nil
	}

	if r.ContentLength == 0 {
		return apierrors.ErrBatchUpdateTestPlanCaseRels.MissingParameter("request body").ToResp(), nil
	}
	var req apistructs.TestPlanCaseRelBatchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrBatchUpdateTestPlanCaseRels.InvalidParameter(err).ToResp(), nil
	}
	req.TestPlanID = testPlanID
	req.IdentityInfo = identityInfo

	// 查询测试计划
	tp, err := e.testPlan.Get(req.TestPlanID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	req.ProjectID = tp.ProjectID

	if !req.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  tp.ProjectID,
			Resource: apistructs.TestPlanUsecaseRelResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return errorresp.ErrResp(err)
		}
		if !access.Access {
			return apierrors.ErrBatchUpdateTestPlanCaseRels.AccessDenied().ToResp(), nil
		}
	}

	if err = e.testPlan.BatchUpdateTestPlanCaseRels(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// RemoveTestPlanCaseRelIssueRelations 解除测试计划用例与事件缺陷的关联关系
func (e *Endpoints) RemoveTestPlanCaseRelIssueRelations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	relID, err := strconv.ParseUint(vars[urlPathTestPlanCaseRelID], 10, 64)
	if err != nil {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	if r.ContentLength == 0 {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.MissingParameter("request body").ToResp(), nil
	}
	var req apistructs.TestPlanCaseRelIssueRelationRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.InvalidParameter(err).ToResp(), nil
	}
	req.TestPlanID = testPlanID
	req.TestPlanCaseRelID = relID
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
			Resource: apistructs.TestPlanUsecaseRelResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return errorresp.ErrResp(err)
		}
		if !access.Access {
			return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.AccessDenied().ToResp(), nil
		}
	}

	if err = e.testPlan.RemoveTestPlanCaseRelIssueRelations(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// AddTestPlanCaseRelIssueRelations 测试计划用例与事件缺陷 新增关联
func (e *Endpoints) AddTestPlanCaseRelIssueRelations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrAddTestPlanCaseRelIssueRelation.NotLogin().ToResp(), nil
	}

	testPlanID, err := strconv.ParseUint(vars[urlPathTestPlanID], 10, 64)
	if err != nil {
		return apierrors.ErrAddTestPlanCaseRelIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	relID, err := strconv.ParseUint(vars[urlPathTestPlanCaseRelID], 10, 64)
	if err != nil {
		return apierrors.ErrAddTestPlanCaseRelIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	if r.ContentLength == 0 {
		return apierrors.ErrAddTestPlanCaseRelIssueRelation.MissingParameter("request body").ToResp(), nil
	}
	var req apistructs.TestPlanCaseRelIssueRelationAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrAddTestPlanCaseRelIssueRelation.InvalidParameter(err).ToResp(), nil
	}
	req.TestPlanID = testPlanID
	req.TestPlanCaseRelID = relID
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
			Resource: apistructs.TestPlanUsecaseRelResource,
			Action:   apistructs.UpdateAction,
		})
		if err != nil {
			return errorresp.ErrResp(err)
		}
		if !access.Access {
			return apierrors.ErrAddTestPlanCaseRelIssueRelation.AccessDenied().ToResp(), nil
		}
	}

	if err = e.testPlan.AddTestPlanCaseRelIssueRelations(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// InternalRemoveTestPlanCaseRelIssueRelations 仅供内部使用，删除测试计划用例下的 bug 关联关系
func (e *Endpoints) InternalRemoveTestPlanCaseRelIssueRelations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.NotLogin().ToResp(), nil
	}
	// 只允许内部调用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.AccessDenied().ToResp(), nil
	}

	// delete by issue
	issueIDStr := r.URL.Query().Get("issueID")
	if issueIDStr == "" {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.MissingParameter("issueID").ToResp(), nil
	}
	issueID, err := strconv.ParseUint(issueIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrRemoveTestPlanCaseRelIssueRelation.InvalidParameter(fmt.Errorf("invalid issueID: %s", issueIDStr)).ToResp(), nil
	}

	if err := e.testPlan.InternalRemoveTestPlanCaseRelIssueRelationsByIssueID(issueID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}
