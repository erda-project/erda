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

package apim

import (
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var DeleteAPIAssetVersion = apis.ApiSpec{
	Path:         "/api/api-assets/<assetID>/versions/<versionID>",
	BackendPath:  "/api/api-assets/<assetID>/versions/<versionID>",
	Method:       http.MethodDelete,
	Host:         APIMAddr,
	Scheme:       "http",
	CheckLogin:   true,
	Doc:          "删除指定版本的 API 资料",
	RequestType:  nil,
	ResponseType: nil,
}
