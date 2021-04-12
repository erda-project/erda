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

package officer

import "github.com/erda-project/erda/modules/openapi/api/apis"

var OFFICER_GET_PRICE = apis.ApiSpec{
	Path:        "/api/resource/aliyun/price",
	BackendPath: "/api/resource/aliyun/price",
	Method:      "POST",
	Host:        "officer.marathon.l4lb.thisdcos.directory:9029",
	Scheme:      "http",
	CheckLogin:  true,
	Doc:         "summary: 获取阿里云指定资源价格",
}
