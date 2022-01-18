// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package block

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CREATE_BLOCK = apis.ApiSpec{
	Path:        "/api/tmc/dashboard/blocks",
	BackendPath: "/api/msp/dashboard/blocks",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 创建自定义大盘",
	Audit:       auditOperatorBlock(apistructs.AddServiceDashboard),
}

func auditOperatorBlock(tmp apistructs.TemplateName) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		var requestBody struct {
			Name string `json:"name"`
		}
		if err := ctx.BindRequestData(&requestBody); err != nil {
			return err
		}
		projectIdStr := ctx.BindResponseHeader("erda-projectId")
		projectId, err := strconv.Atoi(projectIdStr)
		if err != nil {
			return err
		}
		projectName := ctx.BindResponseHeader("erda-projectName")
		workspace := ctx.BindResponseHeader("erda-workspace")

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      uint64(projectId),
			ProjectID:    uint64(projectId),
			TemplateName: tmp,
			Context: map[string]interface{}{
				"projectName":   projectName,
				"dashboardName": requestBody.Name,
				"workspace":     workspace,
			},
		})
	}
}
