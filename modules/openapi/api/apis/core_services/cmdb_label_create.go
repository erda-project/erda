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

var CMDB_LABEL_CREATE = apis.ApiSpec{
	Path:         "/api/labels",
	BackendPath:  "/api/labels",
	Host:         "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:       "http",
	Method:       http.MethodPost,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.ProjectLabelCreateRequest{},
	ResponseType: apistructs.ProjectLabelCreateResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 创建 label",
	Audit: func(ctx *spec.AuditContext) error {
		var reqBody apistructs.ProjectLabelCreateRequest
		if err := ctx.BindRequestData(&reqBody); err != nil {
			return err
		}
		project, err := ctx.Bundle.GetProject(reqBody.ProjectID)
		if err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      project.ID,
			ProjectID:    project.ID,
			TemplateName: apistructs.CreateProjectLabelTemplate,
			Context:      map[string]interface{}{"label": reqBody.Name, "projectName": project.Name},
		})
	},
}
