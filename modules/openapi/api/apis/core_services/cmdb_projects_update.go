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

package core_services

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

/**
already migration
*/
var CMDB_PROJECT_UPDATE = apis.ApiSpec{
	Path:         "/api/projects/<projectId>",
	BackendPath:  "/api/projects/<projectId>",
	Host:         "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:       "http",
	Method:       "PUT",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.ProjectUpdateRequest{},
	ResponseType: apistructs.ProjectUpdateResponse{},
	Doc:          "summary: 更新项目",
	Audit: func(ctx *spec.AuditContext) error {
		projectID, err := ctx.GetParamInt64("projectId")
		if err != nil {
			return err
		}
		project, err := ctx.GetProject(projectID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      uint64(projectID),
			ProjectID:    uint64(projectID),
			TemplateName: apistructs.UpdateProjectTemplate,
			Context:      map[string]interface{}{"projectName": project.Name},
		})
	},
}
