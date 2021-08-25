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

package autotest

import (
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var QA_REPORTSETS_LIST = apis.ApiSpec{
	Path:        "/api/reportsets/<pipelineID>",
	BackendPath: "/api/reportsets/<pipelineID>",
	Host:        "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:      "http",
	Method:      http.MethodGet,
	CheckLogin:  true,
	Doc:         "QA的上报数据查询",
}
