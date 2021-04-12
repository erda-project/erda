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
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/gittar-adaptor/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// CICDTaskLog 包装 cicd task 获取接口
// dashboard: /api/logs?start=0&end=1576498555732000000&count=-200&stream=stderr&id=pipeline-task-2059&source=job
func (e *Endpoints) CICDTaskLog(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	_, task, err := e.checkTaskPermission(r, vars)
	if err != nil {
		return apierrors.ErrGetCICDTaskLog.InternalError(err).ToResp(), nil
	}

	// 获取日志
	var logReq apistructs.DashboardSpotLogRequest
	if err := queryStringDecoder.Decode(&logReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetCICDTaskLog.InvalidParameter(err).ToResp(), nil
	}
	logReq.ID = task.Extra.UUID
	logReq.Source = apistructs.DashboardSpotLogSourceJob

	log, err := e.bdl.GetLog(logReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(log)
}

// ProxyCICDTaskLogDownload 包装 cicd task 下载接口
func (e *Endpoints) ProxyCICDTaskLogDownload(ctx context.Context, r *http.Request, vars map[string]string) error {
	_, task, err := e.checkTaskPermission(r, vars)
	if err != nil {
		return apierrors.ErrDownloadCICDTaskLog.InternalError(err)
	}

	// proxy
	r.URL.Scheme = "http"
	r.Host = discover.Monitor()
	r.URL.Host = discover.Monitor()
	r.URL.Path = "/api/logs/actions/download"
	q := r.URL.Query()
	q.Add("source", string(apistructs.DashboardSpotLogSourceJob))
	q.Add("id", task.Extra.UUID)
	r.URL.RawQuery = q.Encode()

	return nil
}

func (e *Endpoints) checkTaskPermission(r *http.Request, vars map[string]string) (
	*apistructs.PipelineDetailDTO, *apistructs.PipelineTaskDTO, error) {

	pipelineIDStr := vars["pipelineID"]
	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return nil, nil, errors.Errorf("pipelineID: %s", pipelineIDStr)
	}
	taskIDStr := vars["taskID"]
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		return nil, nil, errors.Errorf("taskID: %s", taskIDStr)
	}

	p, err := e.bdl.GetPipeline(pipelineID)
	if err != nil {
		return nil, nil, err
	}
	task, err := e.bdl.GetPipelineTask(pipelineID, taskID)
	if err != nil {
		return nil, nil, err
	}
	if task.PipelineID != p.ID {
		return nil, nil, errors.Errorf("task not belong to pipeline")
	}

	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return nil, nil, err
	}
	if !identityInfo.IsInternalClient() {
		checkResp, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   identityInfo.UserID,
			Scope:    apistructs.AppScope,
			ScopeID:  p.ApplicationID,
			Resource: apistructs.PipelineResource,
			Action:   apistructs.ReadAction,
		})
		if err != nil {
			return nil, nil, err
		}
		if !checkResp.Access {
			return nil, nil, apierrors.ErrGetCICDTaskLog.AccessDenied()
		}
	}

	return p, task, nil
}
