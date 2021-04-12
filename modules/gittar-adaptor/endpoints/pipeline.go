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
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/pipeline"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) pipelineCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var createReq apistructs.PipelineCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		logrus.Errorf("[alert] failed to decode request body: %v", err)
		return apierrors.ErrCreatePipeline.InvalidParameter("request body").ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}
	createReq.UserID = identityInfo.UserID

	if !identityInfo.IsInternalClient() {
		if err := e.permission.CheckBranchAction(identityInfo, strconv.FormatUint(createReq.AppID, 10),
			createReq.Branch, apistructs.OperateAction); err != nil {
			return errorresp.ErrResp(err)
		}
	}

	// 创建pipeline流程
	reqPipeline, err := e.pipeline.ConvertPipelineToV2(&createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 检查是否需要审批, 如果是, 延长部署步骤的 timeout
	app, err := e.bdl.GetApp(createReq.AppID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	branchrule, err := e.bdl.GetBranchWorkspaceConfigByProject(app.ProjectID, createReq.Branch)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if branchrule.NeedApproval {
		pipelineymlStruct, err := pipelineyml.New([]byte(reqPipeline.PipelineYml))
		if err != nil {
			return errorresp.ErrResp(err)
		}
		pipelineymlSpec := pipelineymlStruct.Spec()
		for _, stage := range pipelineymlSpec.Stages {
			for i := range stage.Actions {
				for k := range stage.Actions[i] {
					if k.String() == "dice" {
						stage.Actions[i][k].Timeout = -1
					}
				}
			}
		}
		convertedyml, err := pipelineyml.GenerateYml(pipelineymlSpec)
		if err != nil {
			return errorresp.ErrResp(err)
		}
		reqPipeline.PipelineYml = string(convertedyml)
	}

	resp, err := e.pipeline.CreatePipelineV2(reqPipeline)
	if err != nil {
		logrus.Errorf("create pipeline failed, reqPipeline: %+v, (%+v)", reqPipeline, err)
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(resp)
}

func (e *Endpoints) pipelineList(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var oriReq apistructs.CICDPipelineListRequest
	err := e.queryStringDecoder.Decode(&oriReq, r.URL.Query())
	if err != nil {
		return apierrors.ErrListPipeline.InvalidParameter(err).ToResp(), nil
	}

	var queryParams = make([]string, 0)
	if oriReq.AppID > 0 {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s",
			apistructs.LabelAppID, strconv.FormatUint(oriReq.AppID, 10)))
	}
	if oriReq.Branches != "" {
		for _, b := range strings.Split(oriReq.Branches, ",") {
			queryParams = append(queryParams, fmt.Sprintf("%s=%s",
				apistructs.LabelBranch, b))
		}
	}

	req := apistructs.PipelinePageListRequest{
		PageNum:                    oriReq.PageNum,
		PageSize:                   oriReq.PageSize,
		YmlNames:                   make([]string, 0),
		Sources:                    []apistructs.PipelineSource{apistructs.PipelineSource(oriReq.Sources)},
		Statuses:                   []string{oriReq.Statuses},
		MustMatchLabelsQueryParams: queryParams,
		AllSources:                 true, // TODO: 后续前端需要传具体的 source
	}

	if oriReq.YmlNames != "" {
		for _, yml := range strings.Split(oriReq.YmlNames, ",") {
			req.YmlNames = append(req.YmlNames, yml)
		}
	}

	pageResult, err := e.bdl.PageListPipeline(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(pageResult)
}

func (e *Endpoints) pipelineYmlList(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var req apistructs.CICDPipelineYmlListRequest
	err := e.queryStringDecoder.Decode(&req, r.URL.Query())
	if err != nil {
		return apierrors.ErrListPipelineYml.InvalidParameter(err).ToResp(), nil
	}
	result := pipeline.GetPipelineYmlList(req, e.bdl)
	return httpserver.OkResp(result)
}

func (e *Endpoints) pipelineAppInvokedCombos(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	appIDStr := r.URL.Query().Get(queryParamAppID)
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrListInvokedCombos.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}

	if err := e.permission.CheckAppAction(identityInfo, appID, apistructs.GetAction); err != nil {
		return errorresp.ErrResp(err)
	}

	selected := spec.PipelineCombosReq{
		Branches: strutil.Split(r.URL.Query().Get(queryParamBranches), ",", true),
		Sources:  strutil.Split(r.URL.Query().Get(queryParamSources), ",", true),
		YmlNames: strutil.Split(r.URL.Query().Get(queryParamYmlNames), ",", true),
	}

	combos, err := e.pipeline.AppCombos(appID, &selected)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(combos)
}

func (e *Endpoints) fetchPipelineByAppInfo(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrFetchPipelineByAppInfo.InvalidParameter(err).ToResp(), nil
	}

	branch := r.URL.Query().Get("branch")
	if branch == "" {
		return apierrors.ErrFetchPipelineByAppInfo.MissingParameter("branch").ToResp(), nil
	}
	appIDStr := r.URL.Query().Get("appID")
	if appIDStr == "" {
		return apierrors.ErrFetchPipelineByAppInfo.MissingParameter("appID").ToResp(), nil
	}
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrFetchPipelineByAppInfo.InvalidParameter(err).ToResp(), nil
	}

	if err := e.permission.CheckAppAction(identityInfo, appID, apistructs.GetAction); err != nil {
		return errorresp.ErrResp(err)
	}

	// fetch ymlPath
	combos, err := e.pipeline.AppCombos(appID, nil)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	var ymlName string
	for _, v := range combos {
		if v.Branch == branch {
			for _, item := range v.PagingYmlNames {
				if item != apistructs.DefaultPipelineYmlName {
					ymlName = item
				}
			}
			break
		}
	}

	// fetch pipelineID
	queryParams := make([]string, 0)
	queryParams = append(queryParams, fmt.Sprintf("%s=%d", apistructs.LabelAppID, appID))
	queryParams = append(queryParams, fmt.Sprintf("%s=%s", apistructs.LabelBranch, branch))
	req := apistructs.PipelinePageListRequest{
		PageNum:                    1,
		PageSize:                   1,
		YmlNames:                   []string{ymlName},
		Sources:                    []apistructs.PipelineSource{apistructs.PipelineSourceDice},
		MustMatchLabelsQueryParams: queryParams,
	}
	pageResult, err := e.bdl.PageListPipeline(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if len(pageResult.Pipelines) == 0 {
		return apierrors.ErrFetchPipelineByAppInfo.NotFound().ToResp(), nil
	}
	pipelineID := int(pageResult.Pipelines[0].ID)

	return httpserver.OkResp(pipelineID)
}

// branchWorkspaceMap 获取该应用下所有符合 gitflow 规范的 branch:workspace 映射
func (e *Endpoints) branchWorkspaceMap(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	appIDStr := r.URL.Query().Get(queryParamAppID)
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetBranchWorkspaceMap.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}

	if err := e.permission.CheckAppAction(identityInfo, appID, apistructs.GetAction); err != nil {
		return errorresp.ErrResp(err)
	}

	m, err := e.pipeline.AllValidBranchWorkspaces(appID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(m)
}

func (e *Endpoints) pipelineRun(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}

	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrRunPipeline.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}

	// 根据 pipelineID 获取 pipeline 信息
	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 运行时的入参，不一定需要
	var runRequest apistructs.PipelineRunRequest
	if err := json.NewDecoder(r.Body).Decode(&runRequest); err != nil {
		logrus.Errorf("error to decode runRequest")
	}

	if err := e.permission.CheckBranchAction(identityInfo, strconv.FormatUint(p.ApplicationID, 10),
		p.Branch, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	if err = e.bdl.RunPipeline(apistructs.PipelineRunRequest{
		PipelineID:        pipelineID,
		IdentityInfo:      identityInfo,
		PipelineRunParams: runRequest.PipelineRunParams,
	}); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) pipelineCancel(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrCancelPipeline.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}

	// 根据 pipelineID 获取 pipeline 信息
	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.permission.CheckAppAction(identityInfo, p.ApplicationID, apistructs.GetAction); err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.bdl.CancelPipeline(apistructs.PipelineCancelRequest{
		PipelineID:   pipelineID,
		IdentityInfo: identityInfo,
	}); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// pipelineRerun 重跑整个 pipeline，相当于一个全新的 pipeline，不需要注入上一次的上下文。
func (e *Endpoints) pipelineRerun(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrRerunFailedPipeline.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}

	var rerunReq apistructs.PipelineRerunRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return apierrors.ErrRerunPipeline.InvalidParameter(err).ToResp(), nil
	}
	if string(reqBody) != "" {
		if err := json.Unmarshal(reqBody, &rerunReq); err != nil {
			logrus.Errorf("[alert] failed to decode request body: %v", err)
			return apierrors.ErrRerunPipeline.InvalidParameter("request body").ToResp(), nil
		}
	}

	// 身份校验
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 根据 pipelineID 获取 pipeline 信息
	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.permission.CheckBranchAction(identityInfo, strconv.FormatUint(p.ApplicationID, 10),
		p.Branch, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	rerunReq.PipelineID = pipelineID
	rerunReq.IdentityInfo = identityInfo

	pipelineDto, err := e.bdl.RerunPipeline(rerunReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(pipelineDto)
}

// pipelineRerunFailed 从失败节点开始重试，会注入上下文
func (e *Endpoints) pipelineRerunFailed(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrRerunFailedPipeline.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}

	var rerunFailedReq apistructs.PipelineRerunFailedRequest
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return apierrors.ErrRerunPipeline.InvalidParameter(err).ToResp(), nil
	}
	if string(reqBody) != "" {
		if err := json.Unmarshal(reqBody, &rerunFailedReq); err != nil {
			logrus.Errorf("[alert] failed to decode request body: %v", err)
			return apierrors.ErrRerunPipeline.InvalidParameter("request body").ToResp(), nil
		}
	}

	// 身份校验
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 根据 pipelineID 获取 pipeline 信息，根据 app 做鉴权
	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.permission.CheckAppAction(identityInfo, p.ApplicationID, apistructs.GetAction); err != nil {
		return errorresp.ErrResp(err)
	}

	rerunFailedReq.PipelineID = pipelineID
	rerunFailedReq.IdentityInfo = identityInfo

	pipelineDto, err := e.bdl.RerunFailedPipeline(rerunFailedReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(pipelineDto)
}

func (e *Endpoints) pipelineGetBranchRule(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineBranchRule.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}
	pipelineinfo, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	validbranch, err := e.bdl.GetBranchWorkspaceConfigByProject(pipelineinfo.ProjectID, pipelineinfo.Branch)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	if validbranch == nil || validbranch.Workspace == "" {
		return errorresp.ErrResp(fmt.Errorf("not found branch rule for [%s]", pipelineinfo.Branch))
	}
	return httpserver.OkResp(validbranch)
}

func (e *Endpoints) pipelineOperate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var req apistructs.PipelineOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrOperatePipeline.InvalidParameter(err).ToResp(), nil
	}

	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrOperatePipeline.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}

	// 身份校验
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// 根据 pipelineID 获取 pipeline 信息，根据 app 做鉴权
	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.permission.CheckBranchAction(identityInfo, strconv.FormatUint(p.ApplicationID, 10),
		p.Branch, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.bdl.OperatePipeline(pipelineID, req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) checkrunCreate(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	var gitEvent apistructs.RepoCreateMrEvent
	if r.Body == nil {
		logrus.Errorf("nil body")
		return apierrors.ErrCreateCheckRun.MissingParameter("body").ToResp(), nil
	}
	err := json.NewDecoder(r.Body).Decode(&gitEvent)
	if err != nil {
		logrus.Errorf("failed to decode body, (%+v)", err)
		return apierrors.ErrCreateCheckRun.InvalidParameter(err).ToResp(), nil
	}
	appID, err := strconv.ParseInt(gitEvent.ApplicationID, 10, 64)
	if err != nil {
		logrus.Errorf("failed to decode body, (%+v)", err)
		return apierrors.ErrCreateCheckRun.InvalidParameter(err).ToResp(), nil
	}
	req := apistructs.CICDPipelineYmlListRequest{
		AppID:  appID,
		Branch: gitEvent.Content.SourceBranch,
	}
	result := pipeline.GetPipelineYmlList(req, e.bdl)
	find := false
	for _, each := range result {
		app, err := e.bdl.GetApp(uint64(appID))
		if err != nil {
			return nil, apierrors.ErrGetApp.InternalError(err)
		}
		strPipelineYml, err := e.pipeline.FetchPipelineYml(app.GitRepo, gitEvent.Content.SourceBranch, each)
		if err != nil {
			continue
		}
		pipelineYml, err := pipelineyml.New([]byte(strPipelineYml))
		if err != nil {
			continue
		}
		exist := false
		if !strings.Contains(strPipelineYml, "merge:") {
			continue
		}
		for _, branch := range pipelineYml.Spec().On.Merge.Branches {
			if gitEvent.Content.TargetBranch == branch {
				exist = true
				break
			}
		}

		if !exist {
			continue
		}
		find = true

		// 创建pipeline流程
		reqPipeline := &apistructs.PipelineCreateRequest{
			AppID:              uint64(appID),
			Branch:             gitEvent.Content.SourceBranch,
			Source:             apistructs.PipelineSourceDice,
			PipelineYmlSource:  apistructs.PipelineYmlSourceGittar,
			PipelineYmlContent: strPipelineYml,
			AutoRun:            true,
			UserID:             gitEvent.Content.MergeUserId,
		}
		validBranch, err := e.bdl.GetBranchWorkspaceConfig(uint64(appID), reqPipeline.Branch)
		if err != nil {
			return nil, apierrors.ErrFetchConfigNamespace.InternalError(err)
		}
		workspace := validBranch.Workspace
		v2, err := e.pipeline.ConvertPipelineToV2(reqPipeline)
		if err != nil {
			logrus.Errorf("failed to convert pipeline to V2, (%+v)", err)
			continue
		}
		v2.ForceRun = true
		v2.PipelineYmlName = fmt.Sprintf("%d/%s/%s/%s", reqPipeline.AppID, workspace, gitEvent.Content.SourceBranch, strings.TrimPrefix(each, "/"))
		spew.Dump(v2)
		resPipeline, err := e.pipeline.CreatePipelineV2(v2)
		if err != nil {
			logrus.Errorf("failed to create pipeline, (%+v)", err)
			continue
		}
		// 新建check-run
		request := apistructs.CheckRun{
			MrID:       int64(gitEvent.Content.RepoMergeId),
			Name:       v2.PipelineYmlName,
			PipelineID: strconv.FormatUint(resPipeline.ID, 10),
			Commit:     gitEvent.Content.SourceSha,
		}
		request.Name = gitEvent.Content.SourceBranch + "/" + each
		request.Status = apistructs.CheckRunStatusInProgress
		_, err = e.bdl.CreateCheckRun(appID, request)
		if err != nil {
			continue
		}

		go func() {
			// 轮询获取测试计划执行结果，时间间隔指数增长，衰退上限300s
			l := loop.New(
				loop.WithDeclineRatio(1.5),
				loop.WithDeclineLimit(time.Second*30),
			)
			err = l.Do(func() (bool, error) {
				pipelineResp, err := e.bdl.GetPipeline(resPipeline.ID)
				if err != nil {
					return true, err
				}

				logrus.Infof("Check pipeline result, status: %s", pipelineResp.Status)

				if !pipelineResp.Status.IsEndStatus() {
					return false, fmt.Errorf("is not end")
				}
				if pipelineResp.Status == apistructs.PipelineStatusTimeout {
					request.Result = apistructs.CheckRunResultTimeout
				} else if pipelineResp.Status == apistructs.PipelineStatusStopByUser {
					request.Result = apistructs.CheckRunResultCancelled
				} else if pipelineResp.Status == apistructs.PipelineStatusFailed {
					request.Result = apistructs.CheckRunResultFailure
				} else if pipelineResp.Status == apistructs.PipelineStatusSuccess {
					request.Result = apistructs.CheckRunResultSuccess
				}
				request.Status = apistructs.CheckRunStatusCompleted
				_, err = e.bdl.CreateCheckRun(appID, request)
				if err != nil {
					return true, err
				}
				if pipelineResp.Status != apistructs.PipelineStatusSuccess {
					err := e.bdl.CloseMergeRequest(appID, gitEvent.Content.RepoMergeId)
					if err != nil {
						return true, err
					}
					return true, errors.Errorf("pipeline status error, status: %s", pipelineResp.Status)
				}
				logrus.Infof("Finish to run pipeline, status: %s", pipelineResp.Status)
				return true, nil
			})
		}()
	}
	if !find {
		return apierrors.ErrCreateCheckRun.NotFound().ToResp(), nil
	}
	return httpserver.OkResp(nil)
}
