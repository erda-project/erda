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
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_ITERATION_DELETE = apis.ApiSpec{
	Path:        "/api/iterations/<id>",
	BackendPath: "/api/iterations/<id>",
	Host:        "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:      "http",
	Method:      http.MethodDelete,
	CheckLogin:  true,
	CheckToken:  true,
	IsOpenAPI:   true,
	Doc:         "summary: 删除迭代",
	Audit: func(ctx *spec.AuditContext) error {
		var respBody apistructs.IterationGetResponse
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
			TemplateName: apistructs.DeleteIterationTemplate,
			Context: map[string]interface{}{"projectName": project.Name, "iterationId": respBody.Data.ID,
				"iterationName": respBody.Data.Title},
		})
	},
}
