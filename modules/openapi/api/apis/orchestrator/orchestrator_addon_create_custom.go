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

package orchestrator

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var ORCHESTRATOR_ADDON_CREATE_CUSTOM = apis.ApiSpec{
	Path:        "/api/addons/actions/create-custom",
	BackendPath: "/api/addons/actions/create-custom",
	Host:        "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:      "http",
	Method:      "POST",
	RequestType: apistructs.CustomAddonCreateRequest{},
	CheckLogin:  true,
	CheckToken:  true,
	IsOpenAPI:   true,
	Doc:         `summary: 创建自定义 addon`,
	Audit: func(ctx *spec.AuditContext) error {
		var reqBody apistructs.CustomAddonCreateRequest
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
			TemplateName: apistructs.CreateCustomAddonTemplate,
			Context: map[string]interface{}{
				"addonName":   fmt.Sprintf("%s/%s", reqBody.AddonName, reqBody.Name),
				"projectName": project.Name,
				"projectId":   fmt.Sprintf("%d", project.ID),
			},
		})
	},
}
