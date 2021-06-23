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

var CMDB_ITERATION_CREATE = apis.ApiSpec{
	Path:         "/api/iterations",
	BackendPath:  "/api/iterations",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodPost,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.IterationCreateRequest{},
	ResponseType: apistructs.IterationCreateResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 创建迭代",
	Audit: func(ctx *spec.AuditContext) error {
		var reqBody apistructs.IterationCreateRequest
		if err := ctx.BindRequestData(&reqBody); err != nil {
			return err
		}

		var respBody apistructs.IterationCreateResponse
		if err := ctx.BindResponseData(&respBody); err != nil {
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
			TemplateName: apistructs.CreateIterationTemplate,
			Context: map[string]interface{}{"iterationId": respBody.Data.ID, "iterationName": reqBody.Title,
				"projectName": project.Name},
		})
	},
}
