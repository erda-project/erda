// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dop

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_ORG_DELETE = apis.ApiSpec{
	Path:        "/api/orgs/<orgID>",
	BackendPath: "/api/orgs/<orgID>",
	Host:        "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:      "http",
	Method:      "DELETE",
	CheckLogin:  true,
	CheckToken:  true,
	IsOpenAPI:   true,
	Doc:         "summary: 删除企业",
	Audit: func(ctx *spec.AuditContext) error {
		orgID, err := ctx.GetParamInt64("orgID")
		if err != nil {
			return err
		}
		var response struct {
			Data apistructs.OrgDTO
		}
		if err := ctx.BindResponseData(&response); err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.SysScope,
			ScopeID:      1,
			OrgID:        uint64(orgID),
			TemplateName: apistructs.DeleteOrgTemplate,
			Context:      map[string]interface{}{"orgName": response.Data.Name},
		})
	},
}
