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
already migration
*/

var CMDB_PROJECT_CREATE = apis.ApiSpec{
	Path:         "/api/projects",
	BackendPath:  "/api/projects",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       "POST",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.ProjectCreateRequest{},
	ResponseType: apistructs.ProjectCreateResponse{},
	Doc:          "summary: 创建项目",
	Audit: func(ctx *spec.AuditContext) error {
		var requestBody apistructs.ProjectCreateRequest
		if err := ctx.BindRequestData(&requestBody); err != nil {
			return err
		}
		var responseBody apistructs.ProjectCreateResponse
		if err := ctx.BindResponseData(&responseBody); err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			ProjectID:    responseBody.Data,
			TemplateName: apistructs.CreateProjectTemplate,
			Context:      map[string]interface{}{"projectName": requestBody.Name},
		})
	},
}
