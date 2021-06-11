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

package dop

import (
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ADAPTOR_CICD_FETCH_PIPELINE_ID = apis.ApiSpec{
	Path:        "/api/cicds/actions/fetch-pipeline-id",
	BackendPath: "/api/cicds/actions/fetch-pipeline-id",
	Host:        "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:      "http",
	Method:      http.MethodGet,
	IsOpenAPI:   true,
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 根据 branch & appID 获取 pipelineID",
}
