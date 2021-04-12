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

package qa

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var QA_SONAR_METRIC_RULES_QUERY_LIST = apis.ApiSpec{
	Path:        "/api/sonar-metric-rules/actions/query-list",
	BackendPath: "/api/sonar-metric-rules/actions/query-list",
	Host:        "qa.marathon.l4lb.thisdcos.directory:3033",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 查询 sonar 的配置列表",
	RequestType: apistructs.SonarMetricRulesListRequest{},
	IsOpenAPI:   true,
}
