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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateTestPlanV2 Create test plan
func (e *Endpoints) CreateTestPlanV2(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateTestPlan.NotLogin().ToResp(), nil
	}

	var req apistructs.TestPlanV2CreateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrCreateTestPlan.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateTestPlan.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	if !req.IsInternalClient() {
		// Authorize
		access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   req.UserID,
			Scope:    apistructs.ProjectScope,
			ScopeID:  req.ProjectID,
			Resource: apistructs.TestPlanV2Resource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return apierrors.ErrCreateTestPlan.InternalError(err).ToResp(), nil
		}
		if !access.Access {
			return apierrors.ErrCreateTestPlan.AccessDenied().ToResp(), nil
		}
	}

	// 请求检查
	if err := req.Check(); err != nil {
		return apierrors.ErrCreateTestPlan.InvalidParameter(err).ToResp(), nil
	}

	// TODO: 检查测试空间是否存在

	testPlanID, err := e.autotestV2.CreateTestPlanV2(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testPlanID)
}

// DeleteTestPlanV2 Delete test plan
func (e *Endpoints) DeleteTestPlanV2(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteTestPlan.NotLogin().ToResp(), nil
	}

	testPlanID, err := getTestPlanID(vars)
	if err != nil {
		return apierrors.ErrDeleteTestPlan.InvalidParameter(err).ToResp(), nil
	}

	if err := e.autotestV2.DeleteTestPlanV2(testPlanID, identityInfo); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testPlanID)
}

// UpdateTestPlanV2 Update test plan
func (e *Endpoints) UpdateTestPlanV2(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateTestPlan.NotLogin().ToResp(), nil
	}

	testPlanID, err := getTestPlanID(vars)
	if err != nil {
		return apierrors.ErrUpdateTestPlan.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanV2UpdateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateTestPlan.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateTestPlan.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	req.TestPlanID = testPlanID

	if err := e.autotestV2.UpdateTestPlanV2(&req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(testPlanID)
}

// PagingTestPlansV2 Page query test plan
func (e *Endpoints) PagingTestPlansV2(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingTestPlans.NotLogin().ToResp(), nil
	}

	var req apistructs.TestPlanV2PagingRequest
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
			Resource: apistructs.TestPlanV2Resource,
			Action:   apistructs.ListAction,
		})
		if err != nil {
			return nil, apierrors.ErrPagingTestPlans.InternalError(err)
		}
		if !access.Access {
			return nil, apierrors.ErrPagingTestPlans.AccessDenied()
		}
	}

	result, err := e.autotestV2.PagingTestPlansV2(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(result, result.UserIDs)
}

// GetTestPlanV2 Get testplan detail
func (e *Endpoints) GetTestPlanV2(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetTestPlan.NotLogin().ToResp(), nil
	}

	testPlanID, err := getTestPlanID(vars)
	if err != nil {
		return apierrors.ErrGetTestPlan.InvalidParameter(err).ToResp(), nil
	}

	testPlan, err := e.autotestV2.GetTestPlanV2(testPlanID, identityInfo)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	userIDs := strutil.DedupSlice([]string{testPlan.Creator, testPlan.Updater})

	return httpserver.OkResp(testPlan, userIDs)
}

// AddTestPlanV2Step Add a step in the test plan
func (e *Endpoints) AddTestPlanV2Step(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrAddTestPlanStep.NotLogin().ToResp(), nil
	}

	testPlanID, err := getTestPlanID(vars)
	if err != nil {
		return apierrors.ErrAddTestPlanStep.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanV2StepAddRequest
	if r.ContentLength == 0 {
		return apierrors.ErrAddTestPlanStep.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrAddTestPlanStep.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	req.TestPlanID = testPlanID

	id, err := e.autotestV2.AddTestPlanV2Step(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(id)
}

// DeleteTestPlanV2Step Delete a step in the test plan
func (e *Endpoints) DeleteTestPlanV2Step(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteTestPlanStep.NotLogin().ToResp(), nil
	}

	testPlanID, err := getTestPlanID(vars)
	if err != nil {
		return apierrors.ErrDeleteTestPlanStep.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanV2StepDeleteRequest
	if r.ContentLength == 0 {
		return apierrors.ErrDeleteTestPlanStep.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDeleteTestPlanStep.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	req.TestPlanID = testPlanID

	if err := e.autotestV2.DeleteTestPlanV2Step(&req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp("succ")
}

// UpdateTestPlanV2Step Update the test plan step
func (e *Endpoints) MoveTestPlanV2Step(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateTestPlanStep.NotLogin().ToResp(), nil
	}

	testPlanID, err := getTestPlanID(vars)
	if err != nil {
		return apierrors.ErrUpdateTestPlanStep.InvalidParameter(err).ToResp(), nil
	}

	var req apistructs.TestPlanV2StepUpdateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateTestPlanStep.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateTestPlanStep.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	req.TestPlanID = testPlanID

	if err := e.autotestV2.MoveTestPlanV2Step(&req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp("succ")
}

// UpdateTestPlanV2Step Update the test plan step
func (e *Endpoints) UpdateTestPlanV2Step(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateTestPlanStep.NotLogin().ToResp(), nil
	}

	stepID, err := strconv.ParseUint(vars["stepID"], 10, 64)
	if err != nil {
		return errorresp.ErrResp(errors.New("testPlanID id parse failed"))
	}

	var req apistructs.TestPlanV2StepUpdateRequest
	if r.ContentLength == 0 {
		return apierrors.ErrUpdateTestPlanStep.MissingParameter("request body").ToResp(), nil
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateTestPlanStep.InvalidParameter(err).ToResp(), nil
	}
	req.IdentityInfo = identityInfo
	req.StepID = stepID

	if err := e.autotestV2.UpdateTestPlanV2Step(&req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp("succ")
}

// UpdateTestPlanV2Step Update the test plan step
func (e *Endpoints) GetTestPlanV2Step(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateTestPlanStep.NotLogin().ToResp(), nil
	}

	stepID, err := strconv.ParseUint(vars["stepID"], 10, 64)
	if err != nil {
		return errorresp.ErrResp(errors.New("testPlanID id parse failed"))
	}

	step, err := e.autotestV2.GetTestPlanV2Step(stepID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(step)
}

func getTestPlanID(vars map[string]string) (uint64, error) {
	testplanIDStr := vars["testPlanID"]
	testplanID, err := strconv.ParseUint(testplanIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("testPlanID id parse failed")
	}
	return testplanID, nil
}

func (e *Endpoints) ExecuteDiceAutotestTestPlans(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.AutotestExecuteTestPlansRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetTestPlan.NotLogin().ToResp(), nil
	}
	req.IdentityInfo = identityInfo

	testPlanIDStr := vars["testPlanID"]
	testPlanID, err := strconv.Atoi(testPlanIDStr)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	req.TestPlan.ID = uint64(testPlanID)

	result, err := e.autotestV2.ExecuteDiceAutotestTestPlan(req)
	if err != nil {
		return apierrors.ErrExecuteTestPlanReport.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) CancelDiceAutotestTestPlans(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	var req apistructs.AutotestCancelTestPlansRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAutoTestScene.InvalidParameter(err).ToResp(), nil
	}

	testPlanIDStr := vars["testPlanID"]
	testPlanID, err := strconv.Atoi(testPlanIDStr)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	req.TestPlan.ID = uint64(testPlanID)

	err = e.autotestV2.CancelDiceAutotestTestPlan(req)
	if err != nil {
		return apierrors.ErrCancelTestPlanReport.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("success")
}

func (e *Endpoints) QueryPipelineSnippetYamlV2(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrQueryPipelineSnippetYaml.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrQueryPipelineSnippetYaml.AccessDenied().ToResp(), nil
	}

	// 校验 body 合法性
	if r.ContentLength == 0 {
		return apierrors.ErrQueryPipelineSnippetYaml.InvalidParameter("missing request body").ToResp(), nil
	}
	var req apistructs.SnippetConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrQueryPipelineSnippetYaml.InvalidParameter(err).ToResp(), nil
	}

	switch req.Labels[apistructs.LabelAutotestExecType] {
	case apistructs.SceneSetsAutotestExecType:
		pipelineYml, err := e.autotestV2.QuerySceneSetPipelineSnippetYaml(req)
		if err != nil {
			return errorresp.ErrResp(err)
		}
		return httpserver.OkResp(pipelineYml)
	case apistructs.SceneAutotestExecType:
		pipelineYml, err := e.autotestV2.QueryScenePipelineSnippetYaml(req)
		if err != nil {
			return errorresp.ErrResp(err)
		}
		return httpserver.OkResp(pipelineYml)
	}

	return errorresp.ErrResp(fmt.Errorf("not found snippet pipelineYaml"))
}
