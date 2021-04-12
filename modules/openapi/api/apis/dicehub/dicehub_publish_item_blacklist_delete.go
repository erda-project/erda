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

var DICEHUB_PUBLISH_ITEM_BLACKLIST_DELETE = apis.ApiSpec{
	Path:         "/api/publish-items/<publishItemId>/blacklist/<blacklistId>",
	BackendPath:  "/api/publish-items/<publishItemId>/blacklist/<blacklistId>",
	Host:         "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:       "http",
	Method:       "DELETE",
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	Doc:          `summary: 删除发布内容黑名单`,
	ResponseType: apistructs.PublishItemDeleteBlacklistResponse{},
	Audit: func(ctx *spec.AuditContext) error {

		var resp apistructs.PublishItemDeleteBlacklistResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}
		if resp.Success {
			if resp.Data.UserName == "" {
				if resp.Data.UserID == "" {
					return nil
				} else {
					return ctx.CreateAudit(&apistructs.Audit{
						ScopeType:    apistructs.OrgScope,
						ScopeID:      uint64(ctx.OrgID),
						TemplateName: apistructs.DeletePublishItemsBlacklistTemplate,
						Context:      map[string]interface{}{"removeUser": resp.Data.UserID, "publishItemContent": resp.Data.PublishItemName},
					})
				}
			} else {
				return ctx.CreateAudit(&apistructs.Audit{
					ScopeType:    apistructs.OrgScope,
					ScopeID:      uint64(ctx.OrgID),
					TemplateName: apistructs.DeletePublishItemsBlacklistTemplate,
					Context:      map[string]interface{}{"removeUser": resp.Data.UserName, "publishItemContent": resp.Data.PublishItemName},
				})
			}
		} else {
			return nil
		}
	},
}
