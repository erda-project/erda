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

package checker

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CREATE_CHECKER_V1 = apis.ApiSpec{
	Path:        "/api/projects/<id>/metrics",
	BackendPath: "/api/v1/msp/checkers/projects/<id>/metrics",
	Host:        "msp.marathon.l4lb.thisdcos.directory:8080",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	Doc:         "summary: 新建url监控指标",
	Audit:       auditCreateMicroserviceStatusPageMetric(apistructs.CreateInitiativeMonitor),
}

func auditCreateMicroserviceStatusPageMetric(tmp apistructs.TemplateName) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		var requestBody struct {
			Name string `json:"name"`
		}
		if err := ctx.BindRequestData(&requestBody); err != nil {
			return err
		}
		var projectID uint64
		var projectName string
		pid, err := strconv.ParseUint(ctx.UrlParams["id"], 10, 64)
		if err == nil {
			projectID = pid
			project, err := ctx.Bundle.GetProject(pid)
			if err == nil && project != nil {
				projectName = project.Name
			}
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      projectID,
			ProjectID:    projectID,
			TemplateName: tmp,
			Context: map[string]interface{}{
				"projectId":   projectID,
				"projectName": projectName,
				"metricName":  requestBody.Name,
			},
		})
	}
}

func auditOperateMicroserviceStatusPageMetric(tmp apistructs.TemplateName) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		id := ctx.UrlParams["id"]
		var projectID uint64
		var projectName string
		metricName := id
		info, err := ctx.Bundle.GetMonitorStatusMetricDetails(id)
		if err == nil && info != nil {
			projectID, metricName = uint64(info.ProjectID), info.Name
			project, err := ctx.Bundle.GetProject(projectID)
			if err == nil && project != nil {
				projectName = project.Name
			}
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      projectID,
			ProjectID:    projectID,
			TemplateName: tmp,
			Context: map[string]interface{}{
				"projectId":   projectID,
				"projectName": projectName,
				"metricName":  metricName,
			},
		})
	}
}
