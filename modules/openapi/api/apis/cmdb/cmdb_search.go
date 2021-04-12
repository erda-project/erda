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

package cmdb

import "github.com/erda-project/erda/modules/openapi/api/apis"

var CMDB_CLUSTER_SEARCH = apis.ApiSpec{
	Path:        "/api/cmdb/clusters/<cluster>/search",
	BackendPath: "/api/clusters/<cluster>/search",
	Host:        "cmdb.marathon.l4lb.thisdcos.directory:9093",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	Doc: `
summary: 提供指定集群下的关键字搜索功能
parameters:
  - in: path
    name: cluster
    type: string
    required: true
    description: 集群名或者ID
  - in: body
    name: keyword
    description: search key
    schema:
      type: object
      required:
        - keyword
      properties:
        keyword:
          type: string
          description: keyword 使用冒号进行分割，左key右value
produces:
  - application/json
responses:
  '200':
    description: OK
    schema:
      type: string
      example: 参考 https://yuque.antfin-inc.com/terminus_paas_dev/paas/gosn9b#zngglt
`,
}
