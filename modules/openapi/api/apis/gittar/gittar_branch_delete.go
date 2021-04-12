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

package gittar

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var GITTAR_BRANCH_DELETE = apis.ApiSpec{
	Path:         "/api/gittar/<org>/<repo>/branches/<*>",
	BackendPath:  "/<org>/<repo>/branches/<*>",
	Host:         "gittar.marathon.l4lb.thisdcos.directory:5566",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	IsOpenAPI:    true,
	ResponseType: apistructs.GittarDeleteBranchResponse{},
	Doc:          `summary: 删除分支`,
}
