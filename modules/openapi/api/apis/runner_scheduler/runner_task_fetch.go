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

package runner_scheduler

import (
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var RUNNER_TASK_FETCH = apis.ApiSpec{
	Path:        "/api/runner/fetch-task",
	BackendPath: "/api/runner/fetch-task",
	Host:        "runner-scheduler.marathon.l4lb.thisdcos.directory:9500",
	Scheme:      "http",
	Method:      http.MethodGet,
	IsOpenAPI:   true,
	CheckLogin:  false,
	CheckToken:  false,
}
