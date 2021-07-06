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

package admin

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var ADMIN_NOTICE_UPDATE = apis.ApiSpec{
	Path:         "/api/notices/<id>",
	BackendPath:  "/api/notices/<id>",
	Host:         "admin.marathon.l4lb.thisdcos.directory:9095",
	Scheme:       "http",
	Method:       http.MethodPut,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.NoticeUpdateRequest{},
	ResponseType: apistructs.NoticeUpdateResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 编辑平台公告",
	Audit: func(ctx *spec.AuditContext) error {

		var req apistructs.NoticeUpdateRequest
		err := ctx.BindRequestData(&req)
		if err != nil {
			return err
		}

		var resp apistructs.NoticeUpdateResponse
		err = ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}

		if resp.Success {
			return ctx.CreateAudit(&apistructs.Audit{
				ScopeType:    apistructs.OrgScope,
				ScopeID:      uint64(ctx.OrgID),
				TemplateName: apistructs.UpdateNoticesTemplate,
				Context:      map[string]interface{}{"notices": req.Content},
			})
		} else {
			return nil
		}
	},
}
