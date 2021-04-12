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

var CMDB_CLUSTER_INSTANCES = apis.ApiSpec{
	Path:        "/api/cmdb/clusters/<cluster>/instances",
	BackendPath: "/api/clusters/<cluster>/instances",
	Host:        "cmdb.marathon.l4lb.thisdcos.directory:9093",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	Doc: `
summary: 获取某集群下的批量实例信息
parameters:
  - in: path
    name: cluster
    type: string
    required: true
    description: 集群名或者ID
  - in: query
    name: type
    type: string
    required: true
    enum: [cluster, host, project, application, runtime, service, component, addon]
    description: 指定请求实例的类型
    example: project
  - in: query
    name: host
    type: string
    description: type 等于 host时，若输入 host，则获取指定 host 信息；如果没有输入，则获取整个集群所有的 hosts 信息
  - in: query
    name: project
    type: string
    description: type 等于 project时，若输入 project，则获取指定 project 实例信息；如果没有输入，则获取整个集群所有的 projects 实例信息
  - in: query
    name: application
    type: string
    description: type 等于 application时，若输入 application，则获取指定 application 实例信息；如果没有输入，则获取整个集群所有的 applications 实例信息
  - in: query
    name: runtime
    type: string
    description: type 等于 runtime时，若输入 runtime，则获取指定 runtime 实例信息；如果没有输入，则获取整个集群所有的 runtimes 实例信息
  - in: query
    name: service
    type: string
    description: type 等于 service时，输入 runtime 必要参数。若输入 service，则获取指定 service 实例信息;如果没有输入，则获取指定runtime下所有的 services 实例信息
  - in: query
    name: component
    type: string
    description: type 等于 component时。若输入 component，则获取指定 component 实例信息;如果没有输入，则获取指定集群所有的 component 实例信息
  - in: query
    name: addon
    type: string
    description: type 等于 addon时。若输入 addon，则获取指定 addon 实例信息;如果没有输入，则获取指定集群所有的 addon 实例信息
produces:
  - application/json
responses:
  '200':
    description: OK
    schema:
      type: array
      items:
        type: object
        properties:
          id:
            type: string
            description: 容器ID
          deleted:
            type: boolean
            description: 资源是否被删除
          started_at:
            type: string
            description: 容器启动时间
          cluster_full_name:
            type: string
            description: 集群ID
          host_private_addr:
            type: string
            description: 宿主机内网地址
          ip_addr:
            type: string
            description: 容器IP地址
          image_name:
            type: string
            description: 容器镜像名
          cpu:
            type: number
            format: double
            description: 分配的cpu
          memory:
            type: integer
            format: int64
            description: 分配的内存（字节）
          disk:
            type: integer
            format: int64
            description: 分配的cpu
          dice_org:
            type: string
            description: 所在的组织
          dice_project:
            type: string
            description: 所在的项目
          dice_application:
            type: string
            description: 所在的应用
          dice_runtime:
            type: string
            description: 所在的runtime
          dice_service:
            type: string
            description: 所在的service
          dice_project_name:
            type: string
            description: 所在的项目名称
          dice_application_name:
            type: string
            description: 所在的应用名称
          dice_runtime_name:
            type: string
            description: 所在的runtime名称
          dice_component:
            type: string
            description: 组件名
          dice_addon:
            type: string
            description: 中间件名
          status:
            type: string
            description: 实例状态
          timestamp:
            type: integer
            format: int64
            description: 消息本身的时间戳
`,
}
