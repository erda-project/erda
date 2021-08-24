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

package dicehub

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var DICEHUB_PUBLISH_ITEM_BLACKLIST_CREATE = apis.ApiSpec{
	Path:         "/api/publish-items/<publishItemId>/blacklist",
	BackendPath:  "/api/publish-items/<publishItemId>/blacklist",
	Host:         "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:       "http",
	Method:       "POST",
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.PublishItemUserlistRequest{},
	ResponseType: apistructs.PublishItemAddBlacklistResponse{},
	Doc:          `summary: 添加发布内容黑名单`,

	Audit: func(ctx *spec.AuditContext) error {

		var req apistructs.PublishItemUserlistRequest
		err := ctx.BindRequestData(&req)
		if err != nil {
			return err
		}

		var resp apistructs.PublishItemAddBlacklistResponse
		err = ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}

		if resp.Success {
			if req.UserName == "" {
				return ctx.CreateAudit(&apistructs.Audit{
					ScopeType:    apistructs.OrgScope,
					ScopeID:      uint64(ctx.OrgID),
					TemplateName: apistructs.AddPublishItemsBlacklistTemplate,
					Context:      map[string]interface{}{"addUser": req.UserID, "publishItemContent": resp.Data.Name},
				})
			} else {
				return ctx.CreateAudit(&apistructs.Audit{
					ScopeType:    apistructs.OrgScope,
					ScopeID:      uint64(ctx.OrgID),
					TemplateName: apistructs.AddPublishItemsBlacklistTemplate,
					Context:      map[string]interface{}{"addUser": req.UserName, "publishItemContent": resp.Data.Name},
				})
			}
		} else {
			return nil
		}
	},
}
