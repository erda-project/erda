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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/action-runner-scheduler/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

func (e *Endpoints) CreateRunnerTask(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var request apistructs.CreateRunnerTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return apierrors.ErrCreateRunnerTask.InvalidParameter(err).ToResp(), nil
	}

	id, err := e.runnerTask.CreateRunnerTask(request)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(id)
}

func (e *Endpoints) UpdateRunnerTask(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrUpdateRunnerTask.InvalidParameter(err).ToResp(), nil
	}
	var request apistructs.UpdateRunnerTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return apierrors.ErrUpdateRunnerTask.InvalidParameter(err).ToResp(), nil
	}
	request.ID = id

	err = e.runnerTask.UpdateRunnerTask(&request)
	if err != nil {
		return apierrors.ErrUpdateRunnerTask.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("")
}

func (e *Endpoints) GetRunnerTask(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetRunnerTask.InvalidParameter(err).ToResp(), nil
	}

	task, err := e.runnerTask.GetRunnerTask(id)
	if err != nil {
		return apierrors.ErrGetRunnerTask.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(task)
}

func (e *Endpoints) FetchRunnerTask(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	task, err := e.runnerTask.FetchRunnerTask()
	if err != nil {
		return apierrors.ErrFetchRunnerTask.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(task)
}

func (e *Endpoints) CollectLogs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	source := vars["source"]
	err := e.bundle.CollectLogs(source, r.Body)
	if err != nil {
		return apierrors.ErrCollectRunnerLogs.InvalidParameter(err).ToResp(), nil
	}
	return httpserver.OkResp("")
}
