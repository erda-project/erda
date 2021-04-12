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

package monitor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var MONITOR_ORGS_CUSTOMIZE_ALERT_CREATE = apis.ApiSpec{
	Path:        "/api/orgCenter/customize/alerts",
	BackendPath: "/api/orgs/customize/alerts",
	Host:        "monitor.marathon.l4lb.thisdcos.directory:7096",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 创建企业自定义告警",
	Audit:       auditCreateOrgAlert(apistructs.CreateOrgCustomAlert),
}

func auditOperateOrgCustomAlert(tmp apistructs.TemplateName, act string) func(ctx *spec.AuditContext) error {
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
		if tmp == apistructs.DeleteOrgCustomAlert {
			var respBody struct {
				apistructs.Header
				Data map[string]interface{} `json:"data"`
			}
			err := ctx.BindResponseData(&respBody)
			if err == nil && respBody.Data != nil && respBody.Data["name"] != nil {
				name = fmt.Sprint(respBody.Data["name"])
			}
		} else {
			alert, err := ctx.Bundle.GetMonitorCustomAlertByID(id)
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
