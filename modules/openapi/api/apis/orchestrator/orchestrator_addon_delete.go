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

var ORCHESTRATOR_ADDON_DELETE = apis.ApiSpec{
	Path:         "/api/addons/<addonId>",
	BackendPath:  "/api/addons/<addonId>",
	Host:         "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	ResponseType: apistructs.AddonFetchResponse{},
	Doc:          `summary: 删除 addon`,
	Audit: func(ctx *spec.AuditContext) error {
		var resp apistructs.AddonFetchResponse
		if err := ctx.BindResponseData(&resp); err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      resp.Data.ProjectID,
			ProjectID:    resp.Data.ProjectID,
			TemplateName: apistructs.DeleteAddonTemplate,
			Context: map[string]interface{}{
				"addonName":   fmt.Sprintf("%s/%s", resp.Data.AddonName, resp.Data.Name),
				"projectName": resp.Data.ProjectName,
				"projectId":   resp.Data.ProjectID,
			},
		})
	},
}
