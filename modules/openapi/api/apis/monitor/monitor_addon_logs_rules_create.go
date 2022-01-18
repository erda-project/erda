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

package monitor

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var MONITOR_ADDON_LOGS_RULES_CREATE = apis.ApiSpec{
	Path:        "/api/micro-service/logs/rules",
	BackendPath: "/api/logs/metric/micro_service/rules",
	Host:        "monitor.marathon.l4lb.thisdcos.directory:7096",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 创建日志规则",
	Audit:       auditOperatorBlock(apistructs.CreateAnalyzerRule),
}

func auditOperatorBlock(tmp apistructs.TemplateName) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		var requestBody struct {
			Name string `json:"name"`
		}
		var respBody struct {
			Data string `json:"data"`
		}
		if err := ctx.BindRequestData(&requestBody); err != nil {
			return err
		}
		if err := ctx.BindResponseData(&respBody); err != nil {
			return err
		}
		info, err := ctx.Bundle.GetTenantGroupDetails(ctx.UrlParams["tenantGroup"])
		if err != nil {
			return err
		}
		if len(info.ProjectID) <= 0 {
			return nil
		}
		projectID, err := strconv.ParseUint(info.ProjectID, 10, 64)
		if err != nil {
			return err
		}
		project, err := ctx.Bundle.GetProject(projectID)
		if err != nil {
			return err
		}
		if project == nil {
			return nil
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      projectID,
			ProjectID:    projectID,
			TemplateName: tmp,
			Context: map[string]interface{}{
				"projectName": project.Name,
				"analyzeRule": requestBody.Name,
				"workspace":   respBody.Data,
			},
		})
	}
}
