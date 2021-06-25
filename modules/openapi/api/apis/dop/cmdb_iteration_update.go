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

var CMDB_ITERATION_UPDATE = apis.ApiSpec{
	Path:         "/api/iterations/<id>",
	BackendPath:  "/api/iterations/<id>",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodPut,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.IterationUpdateRequest{},
	ResponseType: apistructs.IterationUpdateResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 更新迭代",
	Audit: func(ctx *spec.AuditContext) error {
		var respBody apistructs.IterationUpdateResponse
		if err := ctx.BindResponseData(&respBody); err != nil {
			return err
		}
		iteration, err := ctx.Bundle.GetIteration(respBody.Data)
		if err != nil {
			return err
		}
		project, err := ctx.Bundle.GetProject(iteration.ProjectID)
		if err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      project.ID,
			ProjectID:    project.ID,
			TemplateName: apistructs.UpdateIterationTemplate,
			Context: map[string]interface{}{"projectName": project.Name, "iterationId": iteration.ID,
				"iterationName": iteration.Title},
		})
	},
}
