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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_LABEL_UPDATE = apis.ApiSpec{
	Path:        "/api/labels/<id>",
	BackendPath: "/api/labels/<id>",
	Host:        "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:      "http",
	Method:      http.MethodPut,
	CheckLogin:  true,
	CheckToken:  true,
	RequestType: apistructs.ProjectLabelUpdateRequest{},
	IsOpenAPI:   true,
	Doc:         "summary: 更新 label",
	Audit: func(ctx *spec.AuditContext) error {
		idStr := ctx.UrlParams["id"]
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return err
		}
		label, err := ctx.Bundle.GetLabel(id)
		if err != nil {
			return err
		}
		project, err := ctx.Bundle.GetProject(label.ProjectID)
		if err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      label.ProjectID,
			ProjectID:    label.ProjectID,
			TemplateName: apistructs.UpdateProjectLabelTemplate,
			Context:      map[string]interface{}{"label": label.Name, "projectName": project.Name},
		})
	},
}
