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

package openapi

import (
	"net/http"

	"github.com/erda-project/erda-infra/base/version"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var OPENAPI_VERSION = apis.ApiSpec{
	Path:   "/api/openapi/version",
	Scheme: "http",
	Method: "GET",
	Custom: func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("short") != "" {
			rw.Write([]byte(version.Version))
			return
		}
		rw.Write([]byte(version.String()))
	},
	Doc: `
summary: 返回 openapi 版本信息
`,
}
