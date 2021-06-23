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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_ORG_UPDATE = apis.ApiSpec{
	Path:         "/api/orgs/<orgID>",
	BackendPath:  "/api/orgs/<orgID>",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       "PUT",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.OrgUpdateRequest{},
	ResponseType: apistructs.OrgUpdateRequestBody{},
	Doc:          "summary: 更新组织",
	Audit: func(ctx *spec.AuditContext) error {
		orgID, err := ctx.GetParamInt64("orgID")
		if err != nil {
			return err
		}
		org, err := ctx.GetOrg(orgID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.SysScope,
			ScopeID:      1,
			OrgID:        uint64(orgID),
			TemplateName: apistructs.UpdateOrgTemplate,
			Context:      map[string]interface{}{"orgName": org.Name},
		})
	},
}
