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
