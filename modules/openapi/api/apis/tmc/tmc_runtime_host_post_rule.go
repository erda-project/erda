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

var TMC_RUNTIME_HOST_POST_RULE = apis.ApiSpec{
	Path:        "/api/tmc/mesh/rule/host/<projectid>/<env>/<host>",
	BackendPath: "/api/tmc/mesh/rule/host/<projectid>/<env>/<host>",
	Host:        "tmc.marathon.l4lb.thisdcos.directory:8050",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	Doc:         "summary: service mesh runtime host rule",
}
