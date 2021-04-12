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

package dicehub

import "github.com/erda-project/erda/modules/openapi/api/apis"

var DICEHUB_RELEASES_PLIST_GET = apis.ApiSpec{
	Path:        "/api/releases/<releaseId>/actions/get-plist",
	BackendPath: "/api/releases/<releaseId>/actions/get-plist",
	Host:        "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:      "http",
	Method:      "GET",
	CheckLogin:  true,
	CheckToken:  true,
	IsOpenAPI:   true,
	Doc:         `获取ios发布类型中的下载 plist 配置`,
}
