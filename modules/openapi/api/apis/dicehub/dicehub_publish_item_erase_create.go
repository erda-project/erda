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

var DICEHUB_PUBLISH_ITEM_ERASE_CREATE = apis.ApiSpec{
	Path:         "/api/publish-items/<publishItemId>/erase",
	BackendPath:  "/api/publish-items/<publishItemId>/erase",
	Host:         "dicehub.marathon.l4lb.thisdcos.directory:10000",
	Scheme:       "http",
	Method:       "POST",
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	Doc:          `summary: 添加发布内容数据擦除`,
	ResponseType: apistructs.PublicItemAddEraseResponse{},
	Audit: func(ctx *spec.AuditContext) error {

		var resp apistructs.PublicItemAddEraseResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}

		if resp.Success {
			return ctx.CreateAudit(&apistructs.Audit{
				ScopeType:    apistructs.OrgScope,
				ScopeID:      uint64(ctx.OrgID),
				TemplateName: apistructs.ErasePublishItemsBlacklistTemplate,
				Context:      map[string]interface{}{"publishItemContent": resp.Data.Data.Name, "eraseUser": resp.Data.DeviceNo},
			})
		} else {
			return nil
		}
		return nil
	},
}
