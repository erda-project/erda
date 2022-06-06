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
	"github.com/erda-project/erda/modules/tools/openapi/legacy/api/apis"
	"github.com/erda-project/erda/modules/tools/openapi/legacy/api/spec"
)

var ADMIN_NOTICE_PUBLISH = apis.ApiSpec{
	Path:         "/api/notices/<id>/actions/publish",
	BackendPath:  "/api/notices/<id>/actions/publish",
	Host:         "erda-server.marathon.l4lb.thisdcos.directory:9095",
	Scheme:       "http",
	Method:       http.MethodPut,
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	Doc:          "summary: 发布平台公告",
	ResponseType: apistructs.NoticePublishResponse{},
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
				TemplateName: apistructs.PublishNoticesTemplate,
				Context:      map[string]interface{}{"notices": resp.Data.Content},
			})
		} else {
			return nil
		}
	},
}
