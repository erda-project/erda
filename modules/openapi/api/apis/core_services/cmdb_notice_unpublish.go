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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_NOTICE_UNPUBLISH = apis.ApiSpec{
	Path:        "/api/notices/<id>/actions/unpublish",
	BackendPath: "/api/notices/<id>/actions/unpublish",
	Host:        "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:      "http",
	Method:      http.MethodPut,
	CheckLogin:  true,
	CheckToken:  true,
	IsOpenAPI:   true,
	Doc:         "summary: 下架平台公告",
	Audit: func(ctx *spec.AuditContext) error {

		var resp apistructs.NoticePublishResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}
		if resp.Success {
			return ctx.CreateAudit(&apistructs.Audit{
				ScopeType:    apistructs.OrgScope,
				ScopeID:      uint64(ctx.OrgID),
				TemplateName: apistructs.UnPublishNoticesTemplate,
				Context:      map[string]interface{}{"notices": resp.Data.Content},
			})
		} else {
			return nil
		}
	},
}
