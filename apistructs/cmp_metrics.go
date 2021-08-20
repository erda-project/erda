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

package apistructs

import (
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	v1 "k8s.io/api/core/v1"
)

type MetricsRequest struct {
	UserID       string
	OrgID        string
	ClusterName  string          `json:"cluster_name"`
	ResourceType v1.ResourceName `json:"resource_type"`
	HostName     []string        `json:"host_name"`
}

type MetricsResponse struct {
	Header
	Data *pb.QueryWithInfluxFormatResponse
}
