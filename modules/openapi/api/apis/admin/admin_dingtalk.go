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

package admin

import (
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ADMIN_DINGTALK_TEST = apis.ApiSpec{
	Path:        "/api/admin/notify/dingtalk-test",
	BackendPath: "/api/admin/notify/dingtalk-test",
	Host:        "admin.marathon.l4lb.thisdcos.directory:9096",
	Scheme:      "http",
	Method:      http.MethodPost,
	CheckLogin:  true,
	CheckToken:  true,
	IsOpenAPI:   true,
	Doc:         "summary: 测试通知组钉钉发送",
}
