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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ADAPTOR_CICD_CONFIG_NAMESPACES = apis.ApiSpec{
	Path:         "/api/cicds/actions/fetch-config-namespaces",
	BackendPath:  "/api/cicds/actions/fetch-config-namespaces",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodGet,
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	ResponseType: apistructs.PipelineConfigNamespacesFetchResponse{},
	Doc:          "summary: 获取应用级别的配置的命名空间",
}
