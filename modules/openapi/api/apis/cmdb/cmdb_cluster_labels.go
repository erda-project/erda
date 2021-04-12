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

var CMDB_CLUSTER_LABELS = apis.ApiSpec{
	Path:         "/api/clusters/actions/statistics-labels",
	BackendPath:  "/api/clusters/actions/statistics-labels",
	Method:       "GET",
	Host:         "cmdb.marathon.l4lb.thisdcos.directory:9093",
	Scheme:       "http",
	CheckLogin:   true,
	RequestType:  apistructs.ClusterLabelsRequest{},
	ResponseType: apistructs.ClusterLabelsResponse{},
	Doc:          "summary: 集群标签占用机器数, 所在机器CPU,MEM总体使用情况, 可调度的CPU,MEM统计等",
}
