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

package adaptor

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ADAPTOR_CICD_BRANCHES_ALL_VALID = apis.ApiSpec{
	Path:         "/api/cicds/actions/app-all-valid-branch-workspaces",
	BackendPath:  "/api/cicds/actions/app-all-valid-branch-workspaces",
	Host:         "gittar-adaptor.marathon.l4lb.thisdcos.directory:1086",
	Scheme:       "http",
	Method:       http.MethodGet,
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	ResponseType: apistructs.PipelineAppAllValidBranchWorkspaceResponse{},
	Doc:          "summary: 获取应用下所有符合 gitflow 规范的分支列表，以及每个分支对应的 workspace",
}
