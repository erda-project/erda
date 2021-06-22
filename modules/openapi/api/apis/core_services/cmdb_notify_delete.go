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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_NOTIFY_DELETE = apis.ApiSpec{
	Path:         "/api/notifies/<notifyID>",
	BackendPath:  "/api/notifies/<notifyID>",
	Host:         "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	ResponseType: apistructs.DeleteNotifyResponse{},
	Doc:          "summary: 删除通知",
	Audit: func(ctx *spec.AuditContext) error {
		var resBody apistructs.DeleteNotifyResponse
		if err := ctx.BindResponseData(&resBody); err != nil {
			return err
		}
		auditReq, err := createNotifyAuditData(ctx, resBody.Data)
		if err != nil {
			return err
		}
		if auditReq.ScopeType == apistructs.ProjectScope {
			auditReq.TemplateName = apistructs.DeleteProjectNotifyTemplate
		} else if auditReq.ScopeType == apistructs.AppScope {
			auditReq.TemplateName = apistructs.DeleteAppNotifyTemplate
		}
		return ctx.CreateAudit(auditReq)
	},
}
