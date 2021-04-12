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

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

// TODO 临时添加，3.8 版本机器详情新增状态字段，监控负责维护，不再走告警工单查询
var CMDB_HOST_STATUS = apis.ApiSpec{
	Path:         "/api/host-status",
	BackendPath:  "/api/resources/host-status",
	Host:         "monitor.marathon.l4lb.thisdcos.directory:7096",
	Scheme:       "http",
	Method:       "POST",
	CheckLogin:   true,
	RequestType:  apistructs.HostStatusListRequest{},
	ResponseType: apistructs.HostStatusListResponse{},
	Doc:          "集群主机状态",
}
