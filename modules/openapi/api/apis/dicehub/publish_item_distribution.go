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

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var PUBLISH_ITEM_DISTRIBUTION = apis.ApiSpec{
	Path:         "/api/publish-items/<publishItemId>/distribution",
	BackendPath:  "/api/publish-items/<publishItemId>/distribution",
	Host:         "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   false,
	CheckToken:   false,
	ResponseType: apistructs.PublishItemDistributionResponse{},
	Doc:          "summary: 发布下载信息",
}
