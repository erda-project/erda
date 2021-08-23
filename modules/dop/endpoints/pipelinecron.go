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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

func (e *Endpoints) pipelineCronPaging(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var req apistructs.PipelineCronPagingRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingPipelineCron.InvalidParameter(err).ToResp(), nil
	}

	crons, err := e.bdl.PageListPipelineCrons(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(crons)
}

func (e *Endpoints) pipelineCronStart(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	cronID, err := strconv.ParseUint(vars[pathCronID], 10, 64)
	if err != nil {
		return apierrors.ErrStartPipelineCron.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}

	// get cron info for check permission
	cronInfo, err := e.bdl.GetPipelineCron(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.permission.CheckRuntimeBranch(identityInfo, cronInfo.ApplicationID, cronInfo.Branch, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	cron, err := e.bdl.StartPipelineCron(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(cron)
}

func (e *Endpoints) pipelineCronStop(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	cronID, err := strconv.ParseUint(vars[pathCronID], 10, 64)
	if err != nil {
		return apierrors.ErrStopPipelineCron.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}

	// get cron info for check permission
	cronInfo, err := e.bdl.GetPipelineCron(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.permission.CheckRuntimeBranch(identityInfo, cronInfo.ApplicationID, cronInfo.Branch, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	cron, err := e.bdl.StopPipelineCron(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(cron)
}

// pipelineCronCreate accept
func (e *Endpoints) pipelineCronCreate(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	var req apistructs.PipelineCronCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreatePipelineCron.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		logrus.Errorf("failed to get identityInfo when create pipeline cron, req: %+v, err: %v", req, err)
		return apierrors.ErrCreatePipelineCron.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrCreatePipelineCron.AccessDenied().ToResp(), nil
	}

	cron, err := e.bdl.CreatePipelineCron(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(cron)
}

func (e *Endpoints) pipelineCronDelete(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	cronID, err := strconv.ParseUint(vars[pathCronID], 10, 64)
	if err != nil {
		return apierrors.ErrDeletePipelineCron.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetUser.InvalidParameter(err).ToResp(), nil
	}

	// get cron info for check permission
	cronInfo, err := e.bdl.GetPipelineCron(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.permission.CheckRuntimeBranch(identityInfo, cronInfo.ApplicationID, cronInfo.Branch, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	if err := e.bdl.DeletePipelineCron(cronID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

// pipelineUpdate pipeline cron update
func (e *Endpoints) pipelineCronUpdate(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.GittarPushPayloadEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdatePipeline.InvalidParameter(err).ToResp(), nil
	}

	if err := e.pipeline.PipelineCronUpdate(req); err != nil {
		return apierrors.ErrUpdatePipeline.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("ok")
}
