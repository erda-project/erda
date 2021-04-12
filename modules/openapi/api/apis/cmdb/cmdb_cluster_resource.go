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

var CMDB_CLUSTER_RESOURCE = apis.ApiSpec{
	Path:         "/api/clusters/actions/accumulate-resource",
	BackendPath:  "/api/clusters/actions/accumulate-resource",
	Method:       "GET",
	Host:         "cmdb.marathon.l4lb.thisdcos.directory:9093",
	Scheme:       "http",
	CheckLogin:   true,
	RequestType:  apistructs.ClusterQueryRequest{},
	ResponseType: apistructs.ClusterResourceResponse{},
	Doc:          "summary: 根据给定集群统计项目，应用，主机，异常主机和runtime的数量",
}
