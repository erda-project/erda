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

package orchestrator

import "github.com/erda-project/erda/modules/openapi/api/apis"

var ORCHESTRATOR_RUNTIME_ROLLBACK = apis.ApiSpec{
	Path:        "/api/runtimes/<runtimeId>/actions/rollback",
	BackendPath: "/api/runtimes/<runtimeId>/actions/rollback",
	Host:        "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	Doc: `
summary: 回滚 Runtime (只能回滚到成功的部署单)
consumes:
  - application/json
parameters:
  - in: path
    name: runtimeId
    type: integer
    required: true
    description: Runtime Id
  - in: body
    description: rollback body
    schema:
      type: object
      properties:
        deploymentId:
          type: int
          description: Deployment Id
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
