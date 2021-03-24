package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/actionagent"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) pipelineTaskDetail(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineTaskDetail.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}

	taskIDStr := vars[pathTaskID]
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineTaskDetail.InvalidParameter(
			strutil.Concat(pathTaskID, ": ", taskIDStr)).ToResp(), nil
	}

	p, err := e.pipelineSvc.Detail(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	task, err := e.pipelineSvc.TaskDetail(taskID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if task.PipelineID != p.ID {
		return apierrors.ErrGetPipelineTaskDetail.InvalidParameter("task not belong to pipeline").ToResp(), nil
	}

	// 校验用户在应用对应分支下是否有 GET 权限
	if err := e.checkBranchPermission(r, p.Labels[apistructs.LabelAppID], p.Labels[apistructs.LabelBranch], apistructs.GetAction); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(task.Convert2DTO())
}

func (e *Endpoints) taskBootstrapInfo(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	pipelineIDStr := vars[pathPipelineID]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetTaskBootstrapInfo.InvalidParameter(
			strutil.Concat(pathPipelineID, ": ", pipelineIDStr)).ToResp(), nil
	}

	taskIDStr := vars[pathTaskID]
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetTaskBootstrapInfo.InvalidParameter(
			strutil.Concat(pathTaskID, ": ", taskIDStr)).ToResp(), nil
	}

	p, err := e.pipelineSvc.Detail(pipelineID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	task, err := e.pipelineSvc.TaskDetail(taskID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if task.PipelineID != p.ID {
		return apierrors.ErrGetTaskBootstrapInfo.InvalidParameter("task not belong to pipeline").ToResp(), nil
	}

	// 只有 action-agent 会使用 token 方式调用该接口，校验已由 openapi checkToken 完成

	// get openapi oauth2 token for callback platform
	_, err = e.pipelineSvc.GetOpenapiOAuth2TokenForActionInvokeOpenapi(task)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	// bootstrapInfoData
	bootstrapInfo := actionagent.AgentArg{
		Commands:    task.Extra.Action.Commands,
		Context:     task.Context,
		PrivateEnvs: task.Extra.PrivateEnvs,
	}
	b, err := json.Marshal(&bootstrapInfo)
	if err != nil {
		return apierrors.ErrGetTaskBootstrapInfo.InternalError(err).ToResp(), nil
	}

	var bootstrapInfoData apistructs.PipelineTaskGetBootstrapInfoResponseData
	bootstrapInfoData.Data = b

	return httpserver.OkResp(bootstrapInfoData)
}
