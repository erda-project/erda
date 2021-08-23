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
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var MONITOR_ORGS_ALERT_CREATE = apis.ApiSpec{
	Path:        "/api/orgCenter/alerts",
	BackendPath: "/api/orgs/alerts",
	Host:        "monitor.marathon.l4lb.thisdcos.directory:7096",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 创建企业告警",
	Audit:       auditCreateOrgAlert(apistructs.CreateOrgAlert),
}

func auditCreateOrgAlert(tmp apistructs.TemplateName) func(ctx *spec.AuditContext) error {
	return func(ctx *spec.AuditContext) error {
		var requestBody struct {
			Name string `json:"name"`
		}
		if err := ctx.BindRequestData(&requestBody); err != nil {
			return err
		}
		org, err := ctx.Bundle.GetOrg(ctx.OrgID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			TemplateName: tmp,
			Context: map[string]interface{}{
				"alertName": requestBody.Name,
				"orgName":   org.Name,
			},
		})
	}
}

func auditOperateOrgAlert(tmp apistructs.TemplateName, act string) func(ctx *spec.AuditContext) error {
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
		org, err := ctx.Bundle.GetOrg(ctx.OrgID)
		if err != nil {
			return err
		}
		if org == nil {
			return nil
		}
		id, err := strconv.ParseInt(ctx.UrlParams["id"], 10, 64)
		if err != nil {
			return err
		}
		name := ctx.UrlParams["id"]
		if tmp == apistructs.DeleteOrgAlert {
			var respBody struct {
				apistructs.Header
				Data map[string]interface{} `json:"data"`
			}
			err := ctx.BindResponseData(&respBody)
			if err == nil && respBody.Data != nil && respBody.Data["name"] != nil {
				name = fmt.Sprint(respBody.Data["name"])
			}
		} else {
			alert, err := ctx.Bundle.GetMonitorAlertByID(id)
			if err == nil && alert != nil {
				name = alert.Name
			}
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			TemplateName: tmp,
			Context: map[string]interface{}{
				"alertID":   id,
				"alertName": name,
				"orgName":   org.Name,
				"action":    action,
			},
		})
	}
}
