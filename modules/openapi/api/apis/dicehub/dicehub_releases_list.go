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

// Package apis dice api集合
package dicehub

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

// DICEHUB_RELEASES_LIST release列表API
var DICEHUB_RELEASES_LIST = apis.ApiSpec{
	Path:         "/api/releases",
	BackendPath:  "/api/releases",
	Host:         "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:       "http",
	Method:       "GET",
	RequestType:  apistructs.ReleaseListRequest{},
	ResponseType: apistructs.ReleaseListResponse{},
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	Doc:          `summary: 版本列表`,
}
