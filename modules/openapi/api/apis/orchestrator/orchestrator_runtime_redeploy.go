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

package orchestrator

import "github.com/erda-project/erda/modules/openapi/api/apis"

var ORCHESTRATOR_RUNTIME_REDEPLOY = apis.ApiSpec{
	Path:        "/api/runtimes/<runtimeId>/actions/redeploy",
	BackendPath: "/api/runtimes/<runtimeId>/actions/redeploy",
	Host:        "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	Doc: `
summary: 重新部署 Runtime (必须要已经部署过一次)
consumes:
  - application/json
parameters:
  - in: path
    name: runtimeId
    type: integer
    required: true
    description: Runtime Id
produces:
  - application/json
responses:
  '200':
    description: ok
    schema:
      type: object
      properties:
        success:
          type: boolean
        err:
          type: object
          properties:
            code:
              type: string
            msg:
              type: string
            ctx:
              type: object
        data:
          type: object
          properties:
            deploymentId:
              type: integer
            applicationId:
              type: integer
            runtimeId:
              type: integer
  '400':
    description: bad request
`,
}
