//  Copyright (c) 2021 Terminus, Inc.
//
//  This program is free software: you can use, redistribute, and/or modify
//  it under the terms of the GNU Affero General Public License, version 3
//  or later ("AGPL"), as published by the Free Software Foundation.
//
//  This program is distributed in the hope that it will be useful, but WITHOUT
//  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
//  FITNESS FOR A PARTICULAR PURPOSE.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program. If not, see <http://www.gnu.org/licenses/>.

package endpoints

import (
	"net/http"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/action-runner-scheduler/services/runnertask"
	"github.com/erda-project/erda/pkg/httpserver"
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
