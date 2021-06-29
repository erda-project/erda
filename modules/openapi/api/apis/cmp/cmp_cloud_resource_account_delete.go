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

package cmp

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMP_CLOUD_RESOURCE_ACCOUNT_DELETE = apis.ApiSpec{
	Path:         "/api/cloud-account",
	BackendPath:  "/api/cloud-account",
	Host:         "cmp.marathon.l4lb.thisdcos.directory:9027",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	RequestType:  apistructs.DeleteCloudAccountRequest{},
	ResponseType: apistructs.DeleteCloudAccountResponse{},
	Doc:          "删除 云账号",
	Audit: func(ctx *spec.AuditContext) error {
		var request apistructs.DeleteCloudAccountRequest
		if err := ctx.BindRequestData(&request); err != nil {
			return err
		}

		if request.Vendor == "" || request.Vendor == "aliyun" {
			request.Vendor = "alicloud"
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			TemplateName: apistructs.DeleteCloudAccountTemplate,
			Context: map[string]interface{}{
				"vendor":      request.Vendor,
				"accessKeyID": request.AccessKey,
			},
		})
	},
}
