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

package core_services

import (
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var CMDB_FILE_HEAD = apis.ApiSpec{
	Path:          "/api/files/<uuid>",
	BackendPath:   "/api/files/<uuid>",
	Host:          "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:        "http",
	Method:        http.MethodHead,
	CheckLogin:    false,
	TryCheckLogin: true,
	CheckToken:    true,
	IsOpenAPI:     true,
	Doc:           "summary: HEAD 请求，文件下载，在 path 中指定具体文件",
}
