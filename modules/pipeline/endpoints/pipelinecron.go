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
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
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

	crons, total, err := e.pipelineCronSvc.Paging(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	var data []*apistructs.PipelineCronDTO
	for _, c := range crons {
		data = append(data, c.Convert2DTO())
	}

	result := apistructs.PipelineCronPagingResponseData{
		Total: total,
		Data:  data,
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) pipelineCronStart(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	cronID, err := strconv.ParseUint(vars[pathCronID], 10, 64)
	if err != nil {
		return apierrors.ErrStartPipelineCron.InvalidParameter(err).ToResp(), nil
	}

	if err := e.checkPipelineCronPermission(r, cronID, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	cron, err := e.pipelineCronSvc.Start(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(cron.Convert2DTO())
}

func (e *Endpoints) pipelineCronStop(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	cronID, err := strconv.ParseUint(vars[pathCronID], 10, 64)
	if err != nil {
		return apierrors.ErrStopPipelineCron.InvalidParameter(err).ToResp(), nil
	}

	if err := e.checkPipelineCronPermission(r, cronID, apistructs.OperateAction); err != nil {
		return errorresp.ErrResp(err)
	}

	cron, err := e.pipelineCronSvc.Stop(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(cron.Convert2DTO())
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

	cron, err := e.pipelineCronSvc.Create(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(cron.ID)
}

func (e *Endpoints) pipelineCronDelete(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	cronID, err := strconv.ParseUint(vars[pathCronID], 10, 64)
	if err != nil {
		return apierrors.ErrDeletePipelineCron.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		logrus.Errorf("failed to get identityInfo when delete pipeline cron, cronID: %d, err: %v", cronID, err)
		return apierrors.ErrDeletePipelineCron.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrDeletePipelineCron.AccessDenied().ToResp(), nil
	}

	if err := e.pipelineCronSvc.Delete(cronID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) pipelineCronGet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	cronID, err := strconv.ParseUint(vars[pathCronID], 10, 64)
	if err != nil {
		return apierrors.ErrGetPipelineCron.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		logrus.Errorf("failed to get identityInfo when get pipeline cron, cronID: %d, err: %v", cronID, err)
		return apierrors.ErrGetPipelineCron.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrGetPipelineCron.AccessDenied().ToResp(), nil
	}

	cron, err := e.pipelineCronSvc.Get(cronID)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(cron.Convert2DTO())
}

// pipelineCronUpdate pipeline cron update
func (e *Endpoints) pipelineCronUpdate(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdatePipelineCron.NotLogin().ToResp(), nil
	}

	var req apistructs.PipelineCronUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdatePipelineCron.InvalidParameter(err).ToResp(), nil
	}

	if err = e.pipelineCronSvc.PipelineCronUpdate(req); err != nil {
		return apierrors.ErrUpdatePipelineCron.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
}
