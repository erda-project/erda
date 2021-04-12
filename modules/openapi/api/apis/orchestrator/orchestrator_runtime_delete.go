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

package orchestrator

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var ORCHESTRATOR_RUNTIME_DELETE = apis.ApiSpec{
	Path:         "/api/runtimes/<runtimeId>",
	BackendPath:  "/api/runtimes/<runtimeId>",
	Host:         "orchestrator.marathon.l4lb.thisdcos.directory:8081",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	CheckToken:   true,
	ResponseType: apistructs.RuntimeDeleteResponse{},
	Doc:          `删除应用实例`,
	Audit: func(ctx *spec.AuditContext) error {
		var resp apistructs.RuntimeDeleteResponse
		if err := ctx.BindResponseData(&resp); err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.AppScope,
			ScopeID:      resp.Data.ApplicationID,
			TemplateName: apistructs.DeleteRuntimeTemplate,
			Context: map[string]interface{}{
				"runtimeName":     resp.Data.Name,
				"applicationName": resp.Data.ApplicationName,
				"workspace":       resp.Data.Workspace,
				"projectId":       fmt.Sprintf("%d", resp.Data.ProjectID),
				"appId":           fmt.Sprintf("%d", resp.Data.ApplicationID),
			},
		})
	},
}
