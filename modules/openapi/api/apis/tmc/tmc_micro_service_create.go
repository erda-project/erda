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

package tmc

import "github.com/erda-project/erda/modules/openapi/api/apis"

var TMC_MICRO_SERVICE_CREATE = apis.ApiSpec{
	Path:        "/api/micro-service/<*>",
	BackendPath: "/api/tmc/micro-service/<*>",
	Host:        "tmc.marathon.l4lb.thisdcos.directory:8050",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 管理微服务的创建",
}
