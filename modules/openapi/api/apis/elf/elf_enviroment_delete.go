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

package elf

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ELF_ENVIROMENT_DELETE = apis.ApiSpec{
	Path:         "/api/v1/environments/<*>",
	BackendPath:  "/api/v1/environments/<*>",
	Host:         "elf.marathon.l4lb.thisdcos.directory:8080",
	Scheme:       "http",
	Method:       http.MethodDelete,
	IsOpenAPI:    true,
	CheckToken:   true,
	ResponseType: apistructs.EnvironmentResponse{},
	Doc:          "summary: 删除环境配置",
}
