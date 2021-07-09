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

package msp

import "github.com/erda-project/erda/modules/openapi/api/apis"

var MSP_EXCEPTION_LIST = apis.ApiSpec{
	Path:        "/api/apm/exceptions",
	BackendPath: "/api/apm/exceptions",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "Query exception list",
}

var MSP_EXCEPTION_EVENT_ID_LIST = apis.ApiSpec{
	Path:        "/api/apm/exception/eventIds",
	BackendPath: "/api/apm/exception/eventIds",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "Query exception event ids",
}

var MSP_EXCEPTION_EVENT = apis.ApiSpec{
	Path:        "/api/apm/exception/event",
	BackendPath: "/api/apm/exception/event",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "Query exception event info",
}
