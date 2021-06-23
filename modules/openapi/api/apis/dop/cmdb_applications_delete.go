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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_APPLICATION_DELETE = apis.ApiSpec{
	Path:         "/api/applications/<applicationId>",
	BackendPath:  "/api/applications/<applicationId>",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       "DELETE",
	IsOpenAPI:    true,
	RequestType:  apistructs.ApplicationDeleteRequest{},
	ResponseType: apistructs.ApplicationDeleteResponse{},
	CheckLogin:   true,
	CheckToken:   true,
	Doc:          "summary: 删除应用",
	Audit: func(ctx *spec.AuditContext) error {
		var resp apistructs.ApplicationDeleteResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}
		applicationDTO := resp.Data
		return ctx.CreateAudit(&apistructs.Audit{
			Context: map[string]interface{}{
				"projectId":   strconv.FormatUint(applicationDTO.ProjectID, 10),
				"appId":       strconv.FormatUint(applicationDTO.ID, 10),
				"projectName": applicationDTO.ProjectName,
				"appName":     applicationDTO.Name,
			},
			AppID:        applicationDTO.ID,
			ProjectID:    applicationDTO.ProjectID,
			ScopeType:    "app",
			ScopeID:      applicationDTO.ID,
			Result:       apistructs.SuccessfulResult,
			TemplateName: apistructs.DeleteAppTemplate,
		})
	},
}
