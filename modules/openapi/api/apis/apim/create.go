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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

const APIMAddr = "apim.marathon.l4lb.thisdcos.directory:3083"

var CreateAPIAsset = apis.ApiSpec{
	Path:         "/api/api-assets",
	BackendPath:  "/api/api-assets",
	Host:         APIMAddr,
	Scheme:       "http",
	Method:       http.MethodPost,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.APIAssetCreateRequest{},
	ResponseType: nil,
	Doc:          "创建 API 资料",
}
