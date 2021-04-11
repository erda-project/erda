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

package cmdb

import (
	"github.com/erda-project/erda/apistructs"

	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_NOTIFYGROUP_DELETE = apis.ApiSpec{
	Path:        "/api/notify-groups/<notifyGroupID>",
	BackendPath: "/api/notify-groups/<notifyGroupID>",
	Host:        "cmdb.marathon.l4lb.thisdcos.directory:9093",
	Scheme:      "http",
	Method:      "DELETE",
	CheckLogin:  true,
	Doc:         "summary: 删除通知组",
	Audit: func(ctx *spec.AuditContext) error {
		var resBody apistructs.DeleteNotifyGroupResponse
		if err := ctx.BindResponseData(&resBody); err != nil {
			return err
		}
		auditReq, err := createNotifyGroupAuditData(ctx, &resBody.Data)
		if err != nil {
			return err
		}
		if auditReq.ScopeType == apistructs.OrgScope {
			auditReq.TemplateName = apistructs.DeleteOrgNotifyGroupTemplate
		} else if auditReq.ScopeType == apistructs.ProjectScope {
			auditReq.TemplateName = apistructs.DeleteProjectNotifyGroupTemplate
		} else if auditReq.ScopeType == apistructs.AppScope {
			auditReq.TemplateName = apistructs.DeleteAppNotifyGroupTemplate
		}
		return ctx.CreateAudit(auditReq)
	},
}
