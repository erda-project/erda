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

package autotest

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var QUERY_PIPELINE_SNIPPET_YAML = apis.ApiSpec{
	Path:        "/api/autotests/pipeline-snippets/actions/query-snippet-yml",
	BackendPath: "/api/autotests/pipeline-snippets/actions/query-snippet-yml",
	Host:        "qa.marathon.l4lb.thisdcos.directory:3033",
	Scheme:      "http",
	Method:      http.MethodPost,
	CheckLogin:  true,
	RequestType: apistructs.AutoTestCaseSavePipelineRequest{},
	Doc:         "查询自动化测试用例流水线文件",
}
