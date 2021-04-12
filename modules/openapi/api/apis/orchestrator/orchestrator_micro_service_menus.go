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

package orchestrator

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ORCHESTRATOR_MICRO_SERVICE_MENUS = apis.ApiSpec{
	Path:         "/api/microservice/project/<projectID>/menus",
	BackendPath:  "/api/microservice/project/<projectID>/menus",
	Host:         "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:       "http",
	Method:       "GET",
	ResponseType: apistructs.MicroServiceMenuResponseData{},
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	Doc:          `summary: 项目对应微服务菜单`,
	Group:        "addons",
}
