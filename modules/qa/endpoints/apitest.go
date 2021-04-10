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
	"net/http/cookiejar"
	"strconv"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/apitestsv2"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"

	"github.com/erda-project/erda/modules/qa/dbclient"
	"github.com/erda-project/erda/modules/qa/services/apierrors"
)

// API 返回对应的错误类型
const (
	ApiTest            = "API_TEST"
	PipelineYmlVersion = "1.1"
	ApiTestType        = "api-test"
	ApiTestIDs         = "api_ids"
	UsecaseID          = "usecase_id"
	PipelineStageLen   = 10
	Project            = "project"
	Usecase            = "case"
)

// CreateAPITest 创建 API 接口测试
func (e *Endpoints) CreateAPITest(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	if r.ContentLength == 0 {
		return apierrors.ErrCreateAPITest.MissingParameter(apierrors.MissingRequestBody).ToResp(), nil
	}
	var req apistructs.ApiTestsCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAPITest.InvalidParameter(err).ToResp(), nil
	}

	api, err := e.testcase.CreateAPI(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(api.ApiID)
}

// UpdateApiTest 更新api接口测试
func (e *Endpoints) UpdateApiTest(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	apiID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateAPITest.InvalidParameter(err).ToResp(), nil
	}

	isResult, _ := strconv.ParseBool(r.URL.Query().Get("isResult"))

	var req apistructs.ApiTestsUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAPITest.InvalidParameter(err).ToResp(), nil
	}

	req.ApiID = apiID
	req.IsResult = isResult

	if _, err := e.testcase.UpdateAPI(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(apiID)
}

// GetApiTests 根据获取接口测试信息
func (e *Endpoints) GetApiTests(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	apiID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrGetAPITest.InvalidParameter(err).ToResp(), nil
	}

	api, err := e.testcase.GetAPI(apiID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(api)
}

// ListApiTests 获取接口测试信息列表
func (e *Endpoints) ListApiTests(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	usecaseID, err := strconv.ParseInt(r.URL.Query().Get("usecaseID"), 10, 64)
	if err != nil {
		return apierrors.ErrListAPITests.InvalidParameter(err).ToResp(), nil
	}

	apiList, err := e.testcase.ListAPIs(usecaseID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(apiList)
}

// DeleteApiTestsByApiID 根据apiID删除接口测试信息
func (e *Endpoints) DeleteApiTestsByApiID(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	apiID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteAPITest.InvalidParameter(err).ToResp(), nil
	}

	err = dbclient.DeleteApiTest(apiID)
	if err != nil {
		return apierrors.ErrDeleteAPITest.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("OK")
}

// ExecuteApiTests 根据planID创建个pipeline执行api测试
func (e *Endpoints) ExecuteApiTests(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	if r.ContentLength == 0 {
		return apierrors.ErrExecuteAPITest.MissingParameter(apierrors.MissingRequestBody).ToResp(), nil
	}
	var req apistructs.ApiTestsActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrExecuteAPITest.InvalidParameter(err).ToResp(), nil
	}

	pipelineID, err := e.testcase.ExecuteAPIs(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(pipelineID)
}

// CancelApiTests 取消执行测试计划
func (e *Endpoints) CancelApiTests(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCancelAPITests.NotLogin().ToResp(), nil
	}

	pipelineID, err := strconv.ParseUint(r.URL.Query().Get("pipelineId"), 10, 64)
	if err != nil {
		return apierrors.ErrCancelAPITests.InvalidParameter(err).ToResp(), nil
	}

	// 根据 pipelineID 获取 pipeline 信息
	if err := e.bdl.CancelPipeline(apistructs.PipelineCancelRequest{
		PipelineID:   pipelineID,
		IdentityInfo: identityInfo,
	}); err != nil {
		return apierrors.ErrCancelAPITests.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("cancel succ")
}

// ExecuteAttemptTest 用户尝试执行单个或者多个API测试
func (e *Endpoints) ExecuteAttemptTest(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	envData := &apistructs.APITestEnvData{
		Header: make(map[string]string),
		Global: make(map[string]*apistructs.APITestEnvVariable),
	}

	var req apistructs.APITestsAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrAttemptExecuteAPITest.InvalidParameter(err).ToResp(), nil
	}

	if len(req.APIs) == 0 {
		return apierrors.ErrAttemptExecuteAPITest.InvalidParameter(fmt.Errorf("API 个数为 0")).ToResp(), nil
	}

	// 获取测试环境变量
	if req.ProjectTestEnvID != 0 {
		envDB, err := dbclient.GetTestEnv(req.ProjectTestEnvID)
		if err != nil || envDB == nil {
			// 忽略错误
			logrus.Warningf("failed to get project test env info, projectID:%d", req.ProjectTestEnvID)
		}

		envData, err = convert2TestEnvResp(envDB)
		if err != nil || envData == nil {
			// 忽略错误
			logrus.Warningf("failed to convert project test env info, env:%+v", envDB)
		}
	}

	if req.UsecaseTestEnvID != 0 {
		envList, err := dbclient.GetTestEnvListByEnvID(req.UsecaseTestEnvID, Usecase)
		if err != nil || envList == nil {
			// 忽略错误
			logrus.Warningf("failed to get usecase test env info, usecaseID:%d", req.UsecaseTestEnvID)
		}

		var envDB dbclient.APITestEnv
		if len(envList) > 0 {
			envDB = envList[0]
			usecaseEnvData, err := convert2TestEnvResp(&envDB)
			if err != nil || usecaseEnvData == nil {
				// 忽略错误
				logrus.Warningf("failed to convert project test env info, env:%+v", envDB)
			}

			if usecaseEnvData != nil {
				// render usecase env data
				if usecaseEnvData.Domain != "" {
					envData.Domain = usecaseEnvData.Domain
				}

				for k, v := range usecaseEnvData.Global {
					envData.Global[k] = v
				}

				for k, v := range usecaseEnvData.Header {
					envData.Header[k] = v
				}
			}
		}
	}

	caseParams := make(map[string]*apistructs.CaseParams)
	// render project env global params, least low priority
	if envData != nil && envData.Global != nil {
		for k, v := range envData.Global {
			caseParams[k] = &apistructs.CaseParams{
				Type:  v.Type,
				Value: v.Value,
			}
		}
	}

	// add cookie jar
	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		logrus.Warningf("failed to new cookie jar")
	}

	httpClient := &http.Client{}

	if cookieJar != nil {
		httpClient.Jar = cookieJar
	}

	respDataList := make([]*apistructs.APITestsAttemptResponseData, 0, len(req.APIs))
	for _, apiInfo := range req.APIs {
		respData := &apistructs.APITestsAttemptResponseData{}
		apiTest := apitestsv2.New(apiInfo, apitestsv2.WithTryV1RenderJsonBodyFirst())
		apiReq, apiResp, err := apiTest.Invoke(httpClient, envData, caseParams)
		if err != nil {
			// 单个 API 执行失败，不返回失败，继续执行下一个
			logrus.Warningf("invoke api error, apiInfo:%+v, (%+v)", apiTest.API, err)
			respData.Response = &apistructs.APIResp{
				BodyStr: err.Error(),
			}
			respData.Request = apiReq
			respDataList = append(respDataList, respData)
			continue
		}
		respData.Response = apiResp
		respData.Request = apiReq

		outParams := apiTest.ParseOutParams(apiTest.API.OutParams, apiResp, caseParams)

		if len(apiTest.API.Asserts) > 0 {
			asserts := apiTest.API.Asserts[0]
			succ, assertResult := apiTest.JudgeAsserts(outParams, asserts)
			logrus.Infof("judge assert result: %v", succ)

			respData.Asserts = &apistructs.APITestsAssertResult{
				Success: succ,
				Result:  assertResult,
			}
		}

		respDataList = append(respDataList, respData)
	}

	return httpserver.OkResp(respDataList)
}

// StatisticResults API 测试结果统计
func (e *Endpoints) StatisticResults(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.APITestsStatisticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrGetStatisticResults.InvalidParameter(err).ToResp(), nil
	}

	statisticResults := &apistructs.APITestsStatisticResponseData{}
	for _, caseID := range req.UsecaseIDs {
		atList, err := dbclient.GetApiTestListByUsecaseID(int64(caseID))
		if err != nil {
			return apierrors.ErrGetStatisticResults.InternalError(err).ToResp(), nil
		}
		statisticResults.Total += uint64(len(atList))
		for _, at := range atList {
			if at.Status == string(apistructs.ApiTestPassed) {
				statisticResults.Passed += 1
			}
		}
	}

	if statisticResults.Total != 0 {
		statisticResults.PassPercent = fmt.Sprintf("%.2f",
			float64(statisticResults.Passed*100)/float64(statisticResults.Total))
	} else {
		statisticResults.PassPercent = "0.00"
	}

	return httpserver.OkResp(statisticResults)
}

// GetPipelineDetail 根据 pipelineID 获取详情
func (e *Endpoints) GetPipelineDetail(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	pipelineID, err := strconv.ParseUint(vars["pipelineID"], 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineDetail.InvalidParameter(err).ToResp(), nil
	}

	respData, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return apierrors.ErrGetPipelineDetail.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(respData)
}

// GetPipelineTaskLogs 根据 taskID 获取 pipeline task 日志详情
func (e *Endpoints) GetPipelineTaskLogs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	pipelineID, err := strconv.ParseUint(vars["pipelineID"], 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineLog.InvalidParameter(err).ToResp(), nil
	}

	taskID, err := strconv.ParseUint(vars["taskID"], 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineLog.InvalidParameter(err).ToResp(), nil
	}

	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return apierrors.ErrGetPipelineLog.InternalError(err).ToResp(), nil
	}

	task, err := e.bdl.GetPipelineTask(pipelineID, taskID)
	if err != nil {
		return apierrors.ErrGetPipelineLog.InternalError(err).ToResp(), nil
	}

	if task.PipelineID != p.ID {
		return apierrors.ErrGetPipelineLog.InternalError(fmt.Errorf("task not belong to pipeline")).ToResp(), nil
	}

	// 获取日志
	var (
		logReq apistructs.DashboardSpotLogRequest
	)

	if err := e.queryStringDecoder.Decode(&logReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetPipelineLog.InternalError(err).ToResp(), nil
	}
	logReq.ID = task.Extra.UUID
	logReq.Source = apistructs.DashboardSpotLogSourceJob

	log, err := e.bdl.GetLog(logReq)
	if err != nil {
		return apierrors.ErrGetPipelineLog.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(*log)
}

func (e *Endpoints) CancelApiTestPipeline(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrTestPlanCancelAPITest.NotLogin().ToResp(), nil
	}
	pipelineID, err := strconv.ParseUint(vars["pipelineID"], 10, 64)
	if err != nil {
		return apierrors.ErrTestPlanCancelAPITest.InvalidParameter(err).ToResp(), nil
	}
	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return apierrors.ErrTestPlanCancelAPITest.InternalError(err).ToResp(), nil
	}
	if p.Source != ApiTestType {
		return apierrors.ErrTestPlanCancelAPITest.AccessDenied().ToResp(), nil
	}
	access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  p.ProjectID,
		Resource: apistructs.TestPlanResource,
		Action:   apistructs.UpdateAction,
	})
	if err != nil {
		return apierrors.ErrTestPlanCancelAPITest.InternalError(err).ToResp(), nil
	}
	if !access.Access {
		return apierrors.ErrTestPlanCancelAPITest.AccessDenied().ToResp(), nil
	}
	err = e.bdl.CancelPipeline(apistructs.PipelineCancelRequest{
		PipelineID:   pipelineID,
		IdentityInfo: identityInfo,
	})
	if err != nil {
		return apierrors.ErrTestPlanCancelAPITest.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("")
}
