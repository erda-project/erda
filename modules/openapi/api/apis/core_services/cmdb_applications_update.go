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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_APPLICATION_UPDATE = apis.ApiSpec{
	Path:         "/api/applications/<applicationId>",
	BackendPath:  "/api/applications/<applicationId>",
	Host:         "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:       "http",
	Method:       "PUT",
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.ApplicationUpdateRequest{},
	ResponseType: apistructs.ApplicationUpdateResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 更新应用",
	Audit: func(ctx *spec.AuditContext) error {
		appID, err := ctx.GetParamInt64("applicationId")
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    "app",
			ScopeID:      uint64(appID),
			TemplateName: apistructs.UpdateAppTemplate,
			Context:      make(map[string]interface{}, 0),
		})
	},
}
