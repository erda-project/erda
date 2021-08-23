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

// Package apis dice api集合
package dicehub

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

// DICEHUB_RELEASE_NAMES_LIST releaseName列表API
var DICEHUB_RELEASE_NAMES_LIST = apis.ApiSpec{
	Path:         "/api/releases/actions/get-name",
	BackendPath:  "/api/releases/actions/get-name",
	Host:         "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:       "http",
	Method:       "GET",
	RequestType:  apistructs.ReleaseNameListRequest{},
	ResponseType: apistructs.ReleaseNameListResponse{},
	IsOpenAPI:    false,
	CheckLogin:   true,
	CheckToken:   true,
	Doc:          `summary: 版本名称列表`,
}
