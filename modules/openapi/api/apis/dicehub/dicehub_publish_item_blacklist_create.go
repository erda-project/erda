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
