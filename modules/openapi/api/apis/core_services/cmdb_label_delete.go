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
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_LABEL_DELETE = apis.ApiSpec{
	Path:        "/api/labels/<id>",
	BackendPath: "/api/labels/<id>",
	Host:        "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:      "http",
	Method:      http.MethodDelete,
	CheckLogin:  true,
	CheckToken:  true,
	IsOpenAPI:   true,
	Doc:         "summary: 删除 label",
	Audit: func(ctx *spec.AuditContext) error {
		var respBody apistructs.ProjectLabelGetByIDResponseData
		if err := ctx.BindResponseData(&respBody); err != nil {
			return err
		}
		project, err := ctx.Bundle.GetProject(respBody.Data.ProjectID)
		if err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      project.ID,
			ProjectID:    project.ID,
			TemplateName: apistructs.DeleteProjectLabelTemplate,
			Context:      map[string]interface{}{"label": respBody.Data.Name, "projectName": project.Name},
		})
	},
}
