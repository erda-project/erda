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

var ADAPTOR_CICD_OPERATE = apis.ApiSpec{
	Path:         "/api/cicds/<pipelineID>",
	BackendPath:  "/api/cicds/<pipelineID>",
	Host:         "gittar-adaptor.marathon.l4lb.thisdcos.directory:1086",
	Scheme:       "http",
	Method:       http.MethodPut,
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.PipelineOperateRequest{},
	ResponseType: apistructs.PipelineOperateResponse{},
	Doc:          "summary: 操作 pipeline",
	Audit: func(ctx *spec.AuditContext) error {
		pipelineId, err := ctx.GetParamUInt64("pipelineID")
		if err != nil {
			return err
		}
		var req apistructs.PipelineOperateRequest
		err = ctx.BindRequestData(&req)
		if err != nil {
			return err
		}
		pipelineDTO, err := ctx.Bundle.GetPipeline(pipelineId)
		if err != nil {
			return err
		}
		taskName := ""
		for _, operate := range req.TaskOperates {
			task, err := ctx.Bundle.GetPipelineTask(pipelineDTO.ID, operate.TaskID)
			if err != nil {
				return err
			}
			taskName = task.Name
		}
		return ctx.CreateAudit(&apistructs.Audit{
			Context: map[string]interface{}{
				"projectId":   strconv.FormatUint(pipelineDTO.ProjectID, 10),
				"appId":       strconv.FormatUint(pipelineDTO.ApplicationID, 10),
				"projectName": pipelineDTO.ProjectName,
				"appName":     pipelineDTO.ApplicationName,
				"pipelineId":  strconv.FormatUint(pipelineDTO.ID, 10),
				"taskName":    taskName,
			},
			ProjectID:    pipelineDTO.ProjectID,
			AppID:        pipelineDTO.ApplicationID,
			ScopeType:    "app",
			TemplateName: apistructs.TogglePipelineTaskTemplate,
			ScopeID:      pipelineDTO.ApplicationID,
		})
	},
}
