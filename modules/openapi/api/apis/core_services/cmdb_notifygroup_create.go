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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_NOTIFYGROUP_CREATE = apis.ApiSpec{
	Path:         "/api/notify-groups",
	BackendPath:  "/api/notify-groups",
	Host:         "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:       "http",
	Method:       "POST",
	CheckLogin:   true,
	RequestType:  apistructs.CreateNotifyGroupRequest{},
	ResponseType: apistructs.CreateNotifyGroupResponse{},
	Doc:          "summary: 创建通知组",
	Audit: func(ctx *spec.AuditContext) error {
		var resBody apistructs.CreateNotifyGroupResponse
		if err := ctx.BindResponseData(&resBody); err != nil {
			return err
		}
		auditReq, err := createNotifyGroupAuditData(ctx, &resBody.Data)
		if err != nil {
			return err
		}
		if auditReq.ScopeType == apistructs.OrgScope {
			auditReq.TemplateName = apistructs.CreateOrgNotifyGroupTemplate
		} else if auditReq.ScopeType == apistructs.ProjectScope {
			auditReq.TemplateName = apistructs.CreateProjectNotifyGroupTemplate
		} else if auditReq.ScopeType == apistructs.AppScope {
			auditReq.TemplateName = apistructs.CreateAppNotifyGroupTemplate
		}
		return ctx.CreateAudit(auditReq)
	},
}

func createNotifyGroupAuditData(ctx *spec.AuditContext, notifyData *apistructs.NotifyGroup) (*apistructs.Audit, error) {
	scopeID, err := strconv.ParseUint(notifyData.ScopeID, 10, 64)
	if err != nil {
		return nil, err
	}
	auditReq := &apistructs.Audit{
		ScopeType: apistructs.ScopeType(notifyData.ScopeType),
		ScopeID:   scopeID,
		Context:   map[string]interface{}{"notifyGroupName": notifyData.Name},
	}
	if auditReq.ScopeType == apistructs.ProjectScope {
		project, err := ctx.GetProject(scopeID)
		if err != nil {
			return nil, err
		}
		auditReq.ProjectID = project.ID
		auditReq.Context["projectName"] = project.Name
		auditReq.Context["projectId"] = project.ID
	} else if auditReq.ScopeType == apistructs.AppScope {
		app, err := ctx.GetApp(scopeID)
		if err != nil {
			return nil, err
		}
		auditReq.AppID = app.ID
		auditReq.ProjectID = app.ProjectID
		auditReq.Context["projectName"] = app.ProjectName
		auditReq.Context["projectId"] = app.ProjectID
		auditReq.Context["appName"] = app.Name
		auditReq.Context["appId"] = app.ID
	}
	return auditReq, nil
}
