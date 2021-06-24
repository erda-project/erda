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

package dop

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var CMDB_ISSUE_IMPORT_EXCEL = apis.ApiSpec{
	Path:         "/api/issues/actions/import-excel",
	BackendPath:  "/api/issues/actions/import-excel",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodPost,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.IssueImportExcelRequest{},
	ResponseType: apistructs.IssueImportExcelResponse{},
	IsOpenAPI:    true,
	Doc:          "导入事件到 excel",
}
