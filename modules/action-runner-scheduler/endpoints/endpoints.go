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
	"net/http"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/action-runner-scheduler/services/runnertask"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

type Endpoints struct {
	runnerTask *runnertask.RunnerTask
	bundle     *bundle.Bundle
}

type Option func(*Endpoints)

// New return an new Endpoints .
func New(options ...Option) *Endpoints {
	e := &Endpoints{}

	for _, op := range options {
		op(e)
	}

	return e
}

// WithRunnerTask set runnerTask .
func WithRunnerTask(runnerTask *runnertask.RunnerTask) Option {
	return func(e *Endpoints) {
		e.runnerTask = runnerTask
	}
}

// WithBundle set bundle .
func WithBundle(bundle *bundle.Bundle) Option {
	return func(e *Endpoints) {
		e.bundle = bundle
	}
}

// Routes return all endpoints methods, i.e. routes.
func (e *Endpoints) Routes() []httpserver.Endpoint {
	return []httpserver.Endpoint{
		{Path: "/healthz", Method: http.MethodGet, Handler: e.Info},

		{Path: "/api/runner/tasks", Method: http.MethodPost, Handler: e.CreateRunnerTask},
		{Path: "/api/runner/tasks/{id}", Method: http.MethodPut, Handler: e.UpdateRunnerTask},
		{Path: "/api/runner/tasks/{id}", Method: http.MethodGet, Handler: e.GetRunnerTask},
		{Path: "/api/runner/fetch-task", Method: http.MethodGet, Handler: e.FetchRunnerTask},
		{Path: "/api/runner/collect/logs/{source}", Method: http.MethodPost, Handler: e.CollectLogs},
	}
}
