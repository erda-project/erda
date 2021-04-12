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
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var GITTAR_TAG_DELETE = apis.ApiSpec{
	Path:         "/api/gittar/<org>/<repo>/tags/<*>",
	BackendPath:  "/<org>/<repo>/tags/<*>",
	Host:         "gittar.marathon.l4lb.thisdcos.directory:5566",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	IsOpenAPI:    true,
	ResponseType: apistructs.GittarDeleteResponse{},
	Doc:          `summary: 删除 tag`,
	Audit: func(ctx *spec.AuditContext) error {
		var responseBody apistructs.GittarDeleteResponse
		if err := ctx.BindResponseData(&responseBody); err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.SysScope,
			ScopeID:      1,
			OrgID:        uint64(ctx.OrgID),
			TemplateName: apistructs.DeleteTagTemplate,
		})
	},
}
