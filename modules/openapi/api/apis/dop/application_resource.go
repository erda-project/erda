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

package dop

import (
	"net/http"
	"net/url"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var APPLICATIONS_RESOURCES_LIST = apis.ApiSpec{
	Path:        "/api/projects/{projectID}/applications-resources",
	BackendPath: "/api/projects/{projectID}/applications-resources",
	Method:      http.MethodGet,
	Host:        "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:      "http",
	CheckLogin:  true,
	CheckToken:  true,
	ChunkAPI:    false,
	Doc:         "the list of applications resources in the project",
	IsOpenAPI:   true,
	Parameters: &apis.Parameters{
		Tag:    "资源相关",
		Header: make(http.Header),
		QueryValues: url.Values{
			"applicationID": nil,
			"ownerID":       nil,
			"orderBy":       []string{"-podsCount,-cpuRequest,-memRequest"},
			"pageNo":        nil,
			"pageSize":      nil,
		},
		Body: nil,
		Response: struct {
			apistructs.Header
			Data interface{} `json:"data"`
		}{
			Header: apistructs.Header{
				Success: true,
				Error:   apistructs.ErrorResponse{},
			},
			Data: apistructs.ApplicationsResourcesResponse{},
		},
	},
}
