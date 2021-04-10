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

var ADAPTOR_CICD_CREATE = apis.ApiSpec{
	Path:         "/api/cicds",
	BackendPath:  "/api/cicds",
	Host:         "gittar-adaptor.marathon.l4lb.thisdcos.directory:1086",
	Scheme:       "http",
	Method:       http.MethodPost,
	IsOpenAPI:    true,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.PipelineCreateRequest{},
	ResponseType: apistructs.PipelineCreateResponse{},
	Doc:          "summary: 创建 pipeline",
	Audit: func(ctx *spec.AuditContext) error {
		var resp apistructs.PipelineCreateResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}
		pipelineDTO := resp.Data
		return ctx.CreateAudit(&apistructs.Audit{
			Context: map[string]interface{}{
				"projectId":   strconv.FormatUint(pipelineDTO.ProjectID, 10),
				"appId":       strconv.FormatUint(pipelineDTO.ApplicationID, 10),
				"projectName": pipelineDTO.ProjectName,
				"appName":     pipelineDTO.ApplicationName,
				"pipelineId":  strconv.FormatUint(pipelineDTO.ID, 10),
				"branch":      pipelineDTO.Branch,
			},
			ProjectID:    pipelineDTO.ProjectID,
			AppID:        pipelineDTO.ApplicationID,
			ScopeType:    "app",
			TemplateName: apistructs.CreatePipelineTemplate,
			ScopeID:      pipelineDTO.ApplicationID,
		})
	},
}
