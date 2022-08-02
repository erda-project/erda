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

package orgapis

import (
	"net/http"

	"github.com/erda-project/erda-infra/providers/httpserver"
)

type MetricSource interface {
	GetContainers(httpserver.Context, *http.Request, struct {
		InstanceType string `param:"instance_type" validate:"required"`
		Start        int64  `query:"start"`
		End          int64  `query:"end"`
	}, resourceRequest) interface{}
	GetHostTypes(*http.Request, struct {
		ClusterName string `query:"clusterName" validate:"required"`
		OrgName     string `query:"orgName" validate:"required"`
	}) interface{}
	GetGroupHosts(*http.Request, struct {
		OrgName string `query:"orgName" validate:"required" json:"-"`
	}, resourceRequest) interface{}
}

type orgChecker interface {
	checkOrgByClusters(ctx httpserver.Context, clusters []*resourceCluster) error
}
