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

var QA_TESTENV_CREATE = apis.ApiSpec{
	Path:         "/api/testenv",
	BackendPath:  "/api/testenv",
	Host:         "qa.marathon.l4lb.thisdcos.directory:3033",
	Scheme:       "http",
	Method:       "POST",
	ResponseType: apistructs.APITestEnvCreateRequest{},
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	Doc:          `summary: 更新项目环境变量信息`,
	Audit: func(ctx *spec.AuditContext) error {
		var req apistructs.APITestEnvCreateRequest
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
			TemplateName: apistructs.QaTestEnvCreateTemplate,
			Context: map[string]interface{}{
				"projectName": project.Name,
				"testEnvName": req.Name,
			},
		})
	},
}
