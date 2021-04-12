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

var PUBLISH_ITEM_VERSION_GET_LATEST = apis.ApiSpec{
	Path:          "/api/publish-items/actions/latest-versions",
	BackendPath:   "/api/publish-items/actions/latest-versions",
	Host:          "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:        "http",
	Method:        "POST",
	CheckLogin:    false,
	TryCheckLogin: true,
	CheckToken:    true,
	IsOpenAPI:     true,
	ChunkAPI:      true,
	RequestType:   apistructs.GetPublishItemLatestVersionRequest{},
	ResponseType:  apistructs.GetPublishItemLatestVersionResponse{},
	Doc:           "summary: 获取移动应用最新的版本",
}
