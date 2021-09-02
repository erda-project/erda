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

package runnertask

import (
	"errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/action-runner-scheduler/conf"
	"github.com/erda-project/erda/modules/action-runner-scheduler/dbclient"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type RunnerTask struct {
	db     *dbclient.DBClient
	bundle *bundle.Bundle
}

// Option RunnerTask config items.
type Option func(*RunnerTask)

// New RunnerTask service
func New(options ...Option) *RunnerTask {
	r := &RunnerTask{}
	for _, op := range options {
		op(r)
	}
	return r
}

// WithDBClient set db client.
func WithDBClient(db *dbclient.DBClient) Option {
	return func(f *RunnerTask) {
		f.db = db
	}
}

// WithBundle set bundle.
func WithBundle(bundle *bundle.Bundle) Option {
	return func(f *RunnerTask) {
		f.bundle = bundle
	}
}

func (f *RunnerTask) CreateRunnerTask(request apistructs.CreateRunnerTaskRequest) (uint64, error) {
	return f.db.CreateRunnerTask(request)
}

func (f *RunnerTask) GetRunnerTask(id int64) (*apistructs.RunnerTask, error) {
	task, err := f.db.GetRunnerTask(id)
	if err != nil {
		return nil, err
	}
	return task.ToApiData(), nil
}

func (f *RunnerTask) UpdateRunnerTask(request *apistructs.UpdateRunnerTaskRequest) error {
	task, err := f.db.GetRunnerTask(request.ID)
	if err != nil {
		return err
	}
	if task.Status != apistructs.RunnerTaskStatusRunning && task.Status != apistructs.RunnerTaskStatusPending {
		return errors.New("invalid task status")
	}
	task.Status = request.Status
	task.ResultDataUrl = request.ResultDataUrl
	return f.db.UpdateRunnerTask(task)
}

func (f *RunnerTask) FetchRunnerTask() ([]*apistructs.RunnerTask, error) {
	task, err := f.db.GetFirstPendingTask()
	if err != nil {
		return nil, err
	}
	if task == nil {
		return []*apistructs.RunnerTask{}, nil
	}

	token, err := f.bundle.GetOpenapiOAuth2Token(apistructs.OpenapiOAuth2TokenGetRequest{
		ClientID:     conf.ClientID(),
		ClientSecret: conf.ClientSecret(),
		Payload: apistructs.OpenapiOAuth2TokenPayload{
			AccessTokenExpiredIn: "3630s",
			AccessibleAPIs: []apistructs.AccessibleAPI{
				{Path: "/api/files", Method: "POST", Schema: "http"},
				{Path: "/api/runner/tasks", Method: "POST", Schema: "http"},
				{Path: "/api/runner/tasks/<runnerTaskID>", Method: "GET", Schema: "http"},
				{Path: "/api/runner/collect/logs/<runnerSource>", Method: "POST", Schema: "http"},
			},
			Metadata: map[string]string{
				httputil.UserHeader: conf.RunnerUserID(),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	task.OpenApiToken = token.AccessToken
	task.Status = apistructs.RunnerTaskStatusRunning
	err = f.db.UpdateRunnerTask(task)
	if err != nil {
		return nil, err
	}

	return []*apistructs.RunnerTask{task.ToApiData()}, nil
}
