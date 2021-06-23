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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

/**
migration
*/
var CMDB_PROJECT_DELETE = apis.ApiSpec{
	Path:         "/api/projects/<projectId>",
	BackendPath:  "/api/projects/<projectId>",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.ProjectDeleteRequest{},
	ResponseType: apistructs.ProjectDeleteResponse{},
	Doc:          "summary: 删除项目",
	Audit: func(ctx *spec.AuditContext) error {
		// 由于与删除project时产生审计事件所需要的返回一样，所以删除project时也用这个接收返回
		var responseBody apistructs.ProjectDetailResponse
		if err := ctx.BindResponseData(&responseBody); err != nil {
			return err
		}
		projectID := responseBody.Data.ID
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      projectID,
			ProjectID:    projectID,
			TemplateName: apistructs.DeleteProjectTemplate,
			Context:      map[string]interface{}{"projectName": responseBody.Data.Name},
		})
	},
}
