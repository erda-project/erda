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

package alert

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
)

const MicroServiceScope = "micro_service"
const CUSTOM_ALERT_TYPE = "micro_service_customize"

const (
	Domain              = "alert_domain"
	DiceOrgId           = "dice_org_id"
	TenantGroup         = "tenant_group"
	TargetPrefix        = "target_"
	ProjectId           = "project_id"
	TargetProjectId     = TargetPrefix + ProjectId
	ApplicationId       = "application_id"
	TargetApplicationId = TargetPrefix + ApplicationId
	WORKSPACE           = "workspace"
	TargetWorkspace     = TargetPrefix + WORKSPACE
	TK                  = "terminus_key"
	TkAlias             = "tk"
	TargetTk            = TargetPrefix + TK
	MetricScope         = "_metric_scope"
	MetricScopeId       = "_metric_scope_id"
	DashboardPath       = "alert_dashboard_path"
	RecordPath          = "alert_record_path"
	Scope               = "alert_scope"
	OperateEq           = "eq"
	OperateIn           = "in"
	OperatorTypeOne     = "one"
	DashboardPathFormat = "/microService/%s/%s/%s/monitor/%s/custom-dashboard"
	RecordPathFormat    = "/microService/%s/%s/%s/monitor/%s/alarm-record"

	NotifyTemplateSample = `【服务HTTP慢事务异常告警】

项目: {{project_name}}

应用: {{application_name}}

服务: {{runtime_name}} - {{service_name}}

事件: {{window}}分钟内HTTP事务平均响应时间{{elapsed_avg_value}} 请求次数{{elapsed_count_sum}}

时间: {{timestamp}}`

	NotifyTemplateSampleEn = `【Service slow http request alarm】

Project: {{project_name}}

Application: {{application_name}}

Service: {{runtime_name}} - {{service_name}}

Event: average response time {{elapsed_avg_value}}

Time: {{timestamp}}"
`
)

func (p *provider) CheckAlert(alert *pb.Alert) interface{} {
	if alert.Name == "" {
		return api.Errors.MissingParameter("alert name")
	}
	if alert.AlertScope == "" {
		return api.Errors.MissingParameter("alert scope")
	}
	if alert.AlertScopeId == "" {
		return api.Errors.MissingParameter("alert scopeId")
	}
	if len(alert.Rules) == 0 {
		return api.Errors.MissingParameter("alert rules")
	}
	if len(alert.Notifies) == 0 {
		return api.Errors.MissingParameter("alert notifies")
	}
	return nil
}

func (p *provider) StringSliceToValue(input []string) (*structpb.Value, error) {
	arr := make([]interface{}, len(input))
	for i, v := range input {
		arr[i] = v
	}
	listValue, err := structpb.NewList(arr)
	if err != nil {
		return nil, err
	}
	return structpb.NewListValue(listValue), nil
}

func (p *provider) GetMicroServiceFilterTags() map[string]bool {
	return p.microServiceFilterTags
}

func (p *provider) checkCustomizeAlert(alert *pb.CustomizeAlertDetail) error {
	if alert.Name == "" {
		return fmt.Errorf("alert name must not be empty")
	}
	if alert.AlertScope == "" {
		return fmt.Errorf("alert scope must not be empty")
	}
	if alert.AlertScopeId == "" {
		return fmt.Errorf("alert scope id must not be empty")
	}
	if len(alert.Rules) == 0 {
		return fmt.Errorf("alert rules id must not be empty")
	}
	if len(alert.Notifies) == 0 {
		return fmt.Errorf("alert notifies must not be empty")
	}
	// 必须包含ticket类型的通知方式，用于告警历史展示
	hasTicket := false
	for _, notify := range alert.Notifies {
		for _, target := range notify.Targets {
			if target == "ticket" {
				hasTicket = true
				break
			}
		}
	}
	if !hasTicket {
		return fmt.Errorf("alert notifies must has ticket")
	}
	return nil
}
