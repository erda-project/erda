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

package adaptor

import (
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var ADAPTOR_CICD_CRON_START = apis.ApiSpec{
	Path:         "/api/cicd-crons/<cronID>/actions/start",
	BackendPath:  "/api/cicd-crons/<cronID>/actions/start",
	Host:         "gittar-adaptor.marathon.l4lb.thisdcos.directory:1086",
	Scheme:       "http",
	Method:       http.MethodPut,
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	ResponseType: apistructs.PipelineCronStartResponse{},
	Doc:          "summary: 开始定时 pipeline",
	Audit: func(ctx *spec.AuditContext) error {
		var res apistructs.PipelineCronStartResponse
		err := ctx.BindResponseData(&res)
		if err != nil {
			return err
		}
		cronDTO := res.Data
		app, err := ctx.GetApp(cronDTO.ApplicationID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			Context: map[string]interface{}{
				"projectId":   strconv.FormatUint(app.ProjectID, 10),
				"appId":       strconv.FormatUint(app.ID, 10),
				"projectName": app.ProjectName,
				"appName":     app.Name,
				"pipelineId":  strconv.FormatUint(cronDTO.BasePipelineID, 10),
				"branch":      cronDTO.Branch,
			},
			ProjectID:    app.ProjectID,
			AppID:        app.ID,
			ScopeType:    "app",
			TemplateName: apistructs.StartPipelineTimerTemplate,
			ScopeID:      app.ID,
		})
	},
}
