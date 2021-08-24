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
