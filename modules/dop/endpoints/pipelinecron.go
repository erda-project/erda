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
