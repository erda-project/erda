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

package member

import "github.com/erda-project/erda/modules/openapi/api/apis"

var MSP_LIST_MEMBER = apis.ApiSpec{
	Path:        "/api/apm/msp/members",
	BackendPath: "/api/apm/msp/members",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "GET",
	CheckToken:  true,
	CheckLogin:  true,
	Doc:         "msp member list",
}
