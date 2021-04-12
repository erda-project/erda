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

package tmc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var TMC_MICRO_SERVICE_ALERT_CREATE = apis.ApiSpec{
	Path:        "/api/tmc/micro-service/tenantGroup/<tenantGroup>/alerts",
	BackendPath: "/api/tmc/micro-service/tenantGroup/<tenantGroup>/alerts",
	Host:        "tmc.marathon.l4lb.thisdcos.directory:8050",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 创建微服务告警",
	Audit:       auditCreateMicroserviceAlert(apistructs.CreateMicroserviceAlert),
}

func auditCreateMicroserviceAlert(tmp apistructs.TemplateName) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		var requestBody struct {
			Name string `json:"name"`
		}
		if err := ctx.BindRequestData(&requestBody); err != nil {
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
			ScopeID:      uint64(projectID),
			ProjectID:    projectID,
			TemplateName: tmp,
			Context: map[string]interface{}{
				"projectId":   projectID,
				"projectName": project.Name,
				"alertName":   requestBody.Name,
			},
		})
	}
}

func auditOperateMicroserviceAlert(tmp apistructs.TemplateName, act string) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		action := act
		if action == "" {
			enable := strings.ToLower(ctx.Request.URL.Query().Get("enable"))
			if enable == "true" {
				action = "enabled"
			} else if enable == "false" {
				action = "disabled"
			}
		}
		tg := ctx.UrlParams["tenantGroup"]
		info, err := ctx.Bundle.GetTenantGroupDetails(tg)
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
		name := "tg=" + tg
		if tmp == apistructs.DeleteMicroserviceAlert {
			var respBody struct {
				apistructs.Header
				Data map[string]interface{} `json:"data"`
			}
			err := ctx.BindResponseData(&respBody)
			if err == nil && respBody.Data != nil && respBody.Data["name"] != nil {
				name = fmt.Sprint(respBody.Data["name"])
			}
		} else {
			data, err := ctx.Bundle.GetMonitorAlertByScope("micro_service", tg)
			if err == nil && data != nil {
				name = data.Name
			}
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      uint64(projectID),
			ProjectID:    projectID,
			TemplateName: tmp,
			Context: map[string]interface{}{
				"projectId":   projectID,
				"projectName": project.Name,
				"alertName":   name,
				"action":      action,
			},
		})
	}
}

func auditOperateMicroserviceCustomAlert(tmp apistructs.TemplateName, act string) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		action := act
		if action == "" {
			enable := strings.ToLower(ctx.Request.URL.Query().Get("enable"))
			if enable == "true" {
				action = "enabled"
			} else if enable == "false" {
				action = "disabled"
			}
		}
		tg := ctx.UrlParams["tenantGroup"]
		info, err := ctx.Bundle.GetTenantGroupDetails(tg)
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
		name := "tg=" + tg
		if tmp == apistructs.DeleteMicroserviceCustomAlert {
			var respBody struct {
				apistructs.Header
				Data map[string]interface{} `json:"data"`
			}
			err := ctx.BindResponseData(&respBody)
			if err == nil && respBody.Data != nil && respBody.Data["name"] != nil {
				name = fmt.Sprint(respBody.Data["name"])
			}
		} else {
			data, err := ctx.Bundle.GetMonitorCustomAlertByScope("micro_service", tg)
			if err == nil && data != nil {
				name = data.Name
			}
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      uint64(projectID),
			ProjectID:    projectID,
			TemplateName: tmp,
			Context: map[string]interface{}{
				"projectId":   projectID,
				"projectName": project.Name,
				"alertName":   name,
				"action":      action,
			},
		})
	}
}
