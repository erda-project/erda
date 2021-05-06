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

package pipeline

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var PIPELINE_BUILD_CACHE_REPORT = apis.ApiSpec{
	Path:         "/api/build-caches",
	BackendPath:  "/api/build-caches",
	Host:         "pipeline.marathon.l4lb.thisdcos.directory:3081",
	Scheme:       "http",
	Method:       "POST",
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.BuildCacheImageReportRequest{},
	ResponseType: apistructs.BuildCacheImageReportResponse{},
	Doc:          "summary: 上报缓存镜像使用情况",
}
