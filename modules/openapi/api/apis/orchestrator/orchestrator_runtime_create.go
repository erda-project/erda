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

var ORCHESTRATOR_RUNTIME_CREATE = apis.ApiSpec{
	Path:        "/api/runtimes",
	BackendPath: "/api/runtimes",
	Host:        "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  false,
	Doc: `
summary: 创建 Runtime
consumes:
  - application/json
parameters:
  - in: body
    description: runtime to create
    schema:
      type: object
      properties:
        name:
          type: string
          description: Runtime Name (pipeline 请用分支名)
        releaseId:
          type: string
          description: diceHub 的 releaseId
        operator:
          type: string
          description: 操作人用户 ID
        clusterName:
          type: string
          description: 告知发布的集群 (aka "az")
        source:
          type: string
          description: PIPELINE / RUNTIMEADDON / ABILITY
        extra:
          type: object
          description: |
            格式为 key(string) / value(object):

                {
                  "k1": "v1",
                  "k2": 123,
                  "k3": ["1", "2", "3"],
                  "k4": {
                    "f1": "g1",
                    "f2": "g2"
                  }
                }

            若为 PIPELINE, 需要传
              - orgId (integer)
              - projectId (integer)
              - applicationId (integer)
              - workspace
              - buildId (integer)

            若为 RUNTIMEADDON, 需要传
              - orgId (integer)
              - projectId (integer)
              - applicationId (integer)
              - workspace
              - instanceId (string)

            若为 ABILITY, 需要传
              - orgId (integer)
              - applicationId (integer) 或 applicationName (自动创建 application)
              - workspace
              - clusterId (string)
              - addonActions (map[string]interface{})
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
