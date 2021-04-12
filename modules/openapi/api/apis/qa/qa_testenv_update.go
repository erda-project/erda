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

package qa

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var QA_TESTENV_UPDATE = apis.ApiSpec{
	Path:         "/api/testenv/<id>",
	BackendPath:  "/api/testenv/<id>",
	Host:         "qa.marathon.l4lb.thisdcos.directory:3033",
	Scheme:       "http",
	Method:       "PUT",
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.APITestEnvUpdateRequest{},
	ResponseType: apistructs.APITestEnvUpdateResponse{},
	IsOpenAPI:    true,
	Doc:          `更新项目环境变量信息`,
	Audit: func(ctx *spec.AuditContext) error {
		var req apistructs.APITestEnvUpdateRequest
		if err := ctx.BindRequestData(&req); err != nil {
			return err
		}
		project, err := ctx.GetProject(req.EnvID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      project.ID,
			TemplateName: apistructs.QaTestEnvUpdateTemplate,
			Context: map[string]interface{}{
				"projectName": project.Name,
				"testEnvName": req.Name,
			},
		})
	},
}
