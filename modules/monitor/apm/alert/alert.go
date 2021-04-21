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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
)

const MicroServiceScope = "micro_service"

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

func (p *provider) queryAlertRule(r *http.Request, params struct {
	ScopeID string `query:"tenantGroup" validate:"required"`
}) interface{} {
	data, err := p.microAlertAPI.QueryAlertRule(r, MicroServiceScope, params.ScopeID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) queryAlert(r *http.Request, params struct {
	ScopeID  string `query:"tenantGroup" validate:"required"`
	PageNo   int    `query:"pageNo" validate:"gte=1"`
	PageSize int    `query:"pageSize" validate:"gte=1,lte=100"`
}) interface{} {
	data, err := p.microAlertAPI.QueryAlert(r, MicroServiceScope, params.ScopeID, uint64(params.PageNo), uint64(params.PageSize))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data == nil {
		data = make([]*adapt.Alert, 0)
	}
	for _, datum := range data {
		datum.Attributes = nil
	}
	total, err := p.microAlertAPI.CountAlert(MicroServiceScope, params.ScopeID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"list":  data,
		"total": total,
	})
}

func (p *provider) getAlert(r *http.Request, params struct {
	ScopeID string `query:"tenantGroup" validate:"required"`
	ID      int    `param:"id" validate:"gte=1"`
}) interface{} {
	data, err := p.microAlertAPI.GetAlertDetail(r, uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	data.Attributes = nil
	if data.AlertScope != MicroServiceScope || data.AlertScopeID != params.ScopeID {
		return api.Errors.AccessDenied()
	}
	return api.Success(data)
}

func (p *provider) getAlertDetail(r *http.Request, params struct {
	ScopeID string `query:"tenantGroup" validate:"required"`
	ID      int    `param:"id" validate:"gte=1"`
}) interface{} {
	data, err := p.microAlertAPI.GetAlertDetail(r, uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(data)
}

func (p *provider) createAlert(r *http.Request, params struct {
	Name           string                   `json:"name"`
	ScopeID        string                   `query:"tenantGroup"`
	ApplicationIds []string                 `json:"appIds"`
	Rules          []*adapt.AlertExpression `json:"rules"`
	Notifies       []*adapt.AlertNotify     `json:"notifies"`
	Domain         string                   `json:"domain"`
	CreateTime     int64                    `json:"createTime"`
	UpdateTime     int64                    `json:"updateTime"`
}) interface{} {

	tk, err := p.authDb.InstanceTenant.QueryTkByTenantGroup(params.ScopeID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	monitorInstance, err := p.authDb.Monitor.GetInstanceByTk(tk)
	if err != nil {
		return api.Errors.Internal(err)
	}

	alert := adapt.Alert{}
	alert.Name = params.Name
	alert.Rules = params.Rules
	alert.Notifies = params.Notifies
	alert.Domain = params.Domain
	alert.CreateTime = params.CreateTime
	alert.UpdateTime = params.UpdateTime
	alert.AlertScope = MicroServiceScope
	alert.AlertScopeID = params.ScopeID
	alert.Attributes = make(map[string]interface{})
	alert.Attributes[DiceOrgId] = monitorInstance.OrgId
	alert.Attributes[Domain] = params.Domain
	alert.Attributes[TenantGroup] = params.ScopeID
	alert.Attributes[ProjectId] = monitorInstance.ProjectId
	alert.Attributes[TargetProjectId] = monitorInstance.ProjectId
	alert.Attributes[WORKSPACE] = monitorInstance.Workspace
	alert.Attributes[TargetWorkspace] = monitorInstance.Workspace
	alert.Attributes[TK] = monitorInstance.TerminusKey
	alert.Attributes[TkAlias] = monitorInstance.TerminusKey
	if params.ApplicationIds != nil && len(params.ApplicationIds) > 0 {
		alert.Attributes[ApplicationId] = params.ApplicationIds
		alert.Attributes[TargetApplicationId] = params.ApplicationIds
	}
	// add dashboard path
	// TODO 修改 dashboard 路径
	alert.Attributes[DashboardPath] =
		fmt.Sprintf(DashboardPathFormat, monitorInstance.ProjectId, monitorInstance.Workspace, params.ScopeID, tk)
	alert.Attributes[RecordPath] =
		fmt.Sprintf(RecordPathFormat, monitorInstance.ProjectId, monitorInstance.Workspace, params.ScopeID, tk)

	if err := p.microAlertAPI.CheckAlert(&alert); err != nil {
		return err
	}
	id, err := p.microAlertAPI.CreateAlert(&alert)
	if err != nil {
		if adapt.IsInvalidParameterError(err) {
			return api.Errors.InvalidParameter(err)
		}
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists(err)
		}
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"id": id,
	})
}

func (p *provider) updateAlert(r *http.Request, params struct {
	ID             int                      `param:"id" validate:"required,gt=0"`
	ScopeID        string                   `query:"tenantGroup" validate:"required"`
	Name           string                   `json:"name"`
	ApplicationIds []string                 `json:"appIds"`
	Enable         bool                     `json:"enable"`
	Rules          []*adapt.AlertExpression `json:"rules"`
	Notifies       []*adapt.AlertNotify     `json:"notifies"`
	Domain         string                   `json:"domain"`
	CreateTime     int64                    `json:"createTime"`
	UpdateTime     int64                    `json:"updateTime"`
}) interface{} {

	data, err := p.microAlertAPI.GetAlert(api.Language(r), uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if MicroServiceScope != data.AlertScope || params.ScopeID != data.AlertScopeID {
		return api.Errors.AccessDenied()
	}

	alert := adapt.Alert{}
	alert.ID = data.ID
	alert.Name = params.Name
	alert.AlertScope = MicroServiceScope
	alert.AlertScopeID = params.ScopeID
	alert.Rules = params.Rules
	alert.Notifies = params.Notifies
	alert.CreateTime = params.CreateTime
	alert.UpdateTime = params.UpdateTime

	if err := p.microAlertAPI.CheckAlert(&alert); err != nil {
		return err
	}
	if err := p.microAlertAPI.UpdateAlert(uint64(params.ID), &alert); err != nil {
		if adapt.IsInvalidParameterError(err) {
			return api.Errors.InvalidParameter(err)
		}
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists(err)
		}
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) updateAlertEnable(r *http.Request, params struct {
	ID      int    `param:"id" validate:"required,gt=0"`
	ScopeID string `query:"tenantGroup" validate:"required"`
	Enable  bool   `query:"enable"`
}) interface{} {
	data, err := p.microAlertAPI.GetAlert(api.Language(r), uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if MicroServiceScope != data.AlertScope || params.ScopeID != data.AlertScopeID {
		return api.Errors.AccessDenied()
	}
	err = p.microAlertAPI.UpdateAlertEnable(uint64(params.ID), params.Enable)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) deleteAlert(r *http.Request, params struct {
	ID      int    `param:"id" validate:"required,gt=0"`
	ScopeID string `query:"tenantGroup" validate:"required"`
}) interface{} {
	data, err := p.microAlertAPI.GetAlert(api.Language(r), uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}

	if MicroServiceScope != data.AlertScope || params.ScopeID != data.AlertScopeID {
		return api.Errors.AccessDenied()
	}

	err = p.microAlertAPI.DeleteAlert(uint64(params.ID))
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data != nil {
		return api.Success(map[string]interface{}{
			"name": data.Name,
		})
	}
	return api.Success(nil)
}

func (p *provider) queryCustomizeMetric(r *http.Request, params struct {
	TenantGroup string `query:"tenantGroup" validate:"required"`
}) interface{} {
	tk, err := p.authDb.InstanceTenant.QueryTkByTenantGroup(params.TenantGroup)
	if err != nil {
		return api.Errors.Internal("get tk failed.")
	}
	data, err := p.microAlertAPI.CustomizeMetrics(api.Language(r), MicroServiceScope, tk, nil)
	if err != nil {
		return api.Errors.Internal("get metrics failed.")
	}
	tags := p.microAlertAPI.GetMicroServiceFilterTags()
	for _, metric := range data.Metrics {
		for i := 0; i < len(metric.Tags); i++ {
			tag := metric.Tags[i]
			if _, ok := tags[tag.Tag.Key]; ok {
				metric.Tags = append(metric.Tags[:i], metric.Tags[i+1:]...)
				i--
			}
		}
	}

	data.NotifySample = NotifyTemplateSample
	lang := api.Language(r)
	if len(lang) > 0 && strings.HasPrefix(strings.ToLower(lang[0].String()), "en") {
		data.NotifySample = NotifyTemplateSampleEn
	}

	if data.FunctionOperators != nil {

	}
	for i := 0; i < len(data.FunctionOperators); i++ {
		operator := data.FunctionOperators[i]
		if operator.Type != OperatorTypeOne {
			data.FunctionOperators = append(data.FunctionOperators[:i], data.FunctionOperators[i+1:]...)
			i--
		}
	}

	for i := 0; i < len(data.FilterOperators); i++ {
		operator := data.FilterOperators[i]
		if operator.Type != OperatorTypeOne {
			data.FilterOperators = append(data.FilterOperators[:i], data.FilterOperators[i+1:]...)
			i--
		}
	}

	return api.Success(data)
}

func (p *provider) queryCustomizeNotifyTarget(params struct {
	ScopeID string `query:"tenantGroup" validate:"required"`
}, r *http.Request) interface{} {
	return api.Success(map[string]interface{}{
		"targets": p.microAlertAPI.NotifyTargetsKeys(api.Language(r), api.OrgID(r)),
	})
}

func (p *provider) queryCustomizeAlerts(r *http.Request, params struct {
	ScopeID  string `query:"tenantGroup" validate:"required"`
	PageNo   int    `query:"pageNo" validate:"gte=1" default:"1"`
	PageSize int    `query:"pageSize" validate:"gte=1,lte=100" default:"20"`
}) interface{} {
	if params.PageSize > 100 {
		params.PageSize = 20
	}

	alert, total, err := p.microAlertAPI.CustomizeAlerts(api.Language(r), MicroServiceScope, params.ScopeID, params.PageNo, params.PageSize)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"total": total,
		"list":  alert,
	})
}

func (p *provider) getCustomizeAlert(params struct {
	ID      uint64 `param:"id" validate:"required,gt=0"`
	ScopeID string `query:"tenantGroup" validate:"required"`
}) interface{} {
	alert, err := p.microAlertAPI.CustomizeAlertDetail(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	if alert.AlertScopeID != params.ScopeID {
		return api.Errors.AccessDenied()
	}
	return api.Success(alert)
}

func (p *provider) getCustomizeAlertDetail(params struct {
	ID uint64 `param:"id" validate:"required,gt=0"`
}) interface{} {
	alert, err := p.microAlertAPI.CustomizeAlertDetail(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(alert)
}

func selectCustomMetrics(customMetric *adapt.CustomizeMetrics) (map[string]*adapt.MetricMeta,
	map[string]*adapt.Operator, map[string]*adapt.Operator, map[string]*adapt.DisplayKey) {

	metricsMap := make(map[string]*adapt.MetricMeta)
	filtersMap := make(map[string]*adapt.Operator)
	functionsMap := make(map[string]*adapt.Operator)
	aggregationsMap := make(map[string]*adapt.DisplayKey)

	// tags
	for _, filter := range customMetric.FilterOperators {
		filtersMap[filter.Key] = filter
	}

	// fields
	for _, field := range customMetric.FunctionOperators {
		functionsMap[field.Key] = field
	}

	// customMetric
	for _, metric := range customMetric.Metrics {
		metricsMap[metric.Name.Key] = metric
	}

	// aggregations
	for _, key := range customMetric.Aggregator {
		aggregationsMap[key.Key] = key
	}

	return metricsMap, filtersMap, functionsMap, aggregationsMap
}

func checkCustomMetric(customMetrics *adapt.CustomizeMetrics, alert adapt.CustomizeAlertDetail) error {

	if customMetrics == nil {
		return errors.New("metric meta not exist")
	}
	metrics, _, functions, aggregations := selectCustomMetrics(customMetrics)

	aliases := make(map[string]string)
	for _, rule := range alert.Rules {
		// check metric
		if metric, ok := metrics[rule.Metric]; !ok {
			return errors.New("rule.Metric not exist")
		} else {
			tagsMap := make(map[string]*adapt.TagMeta)
			fieldsMap := make(map[string]*adapt.FieldMeta)
			rule.Select = make(map[string]string)
			for _, tag := range metric.Tags {
				tagsMap[tag.Tag.Key] = tag
				rule.Select[tag.Tag.Key] = "#" + tag.Tag.Key
			}
			for _, field := range metric.Fields {
				fieldsMap[field.Field.Key] = field
			}

			for _, function := range rule.Functions {
				if function.Alias == "" {
					return errors.New(fmt.Sprintf("the %s function alias is empty", function.Field))
				}
				if alias, ok := aliases[function.Alias]; ok {
					return errors.New(fmt.Sprintf("alias :%s duplicate", alias))
				} else {
					aliases[alias] = function.Alias
				}

				if field, ok := fieldsMap[function.Field]; !ok {
					return errors.New(fmt.Sprintf("not support rule function field %s", field))
				}
				if aggregation, ok := aggregations[function.Aggregator]; !ok {
					return errors.New(fmt.Sprintf("not support rule function aggregator %s", aggregation))
				}
				if operator, ok := functions[function.Operator]; !ok {
					return errors.New(fmt.Sprintf("not support rule function operator %s", operator))
				}

			}

			for _, filter := range rule.Filters {
				if tag, ok := tagsMap[filter.Tag]; !ok {
					return errors.New(fmt.Sprintf("not support rule filter tag %s", tag))
				}
				if operator, ok := functions[filter.Operator]; !ok {
					return errors.New(fmt.Sprintf("not support rule filter operator %s", operator))
				}
			}

			for _, group := range rule.Group {
				if group, ok := tagsMap[group]; !ok {
					return errors.New(fmt.Sprintf("not support rule filter tag %s", group))
				}
			}
			rule.Outputs = append(rule.Outputs, "alert")
		}
	}
	return nil
}

func (p *provider) createCustomizeAlert(params struct {
	TenantGroup string `query:"tenantGroup" validate:"required"`
}, alert adapt.CustomizeAlertDetail, r *http.Request) interface{} {
	tk, err := p.authDb.InstanceTenant.QueryTkByTenantGroup(params.TenantGroup)
	if err != nil {
		return api.Errors.Internal("get tk failed.")
	}
	customizeMetrics, err := p.microAlertAPI.CustomizeMetrics(api.Language(r), MicroServiceScope, tk, nil)
	if err != nil {
		return api.Errors.Internal("get metric meta failed")
	}
	err = checkCustomMetric(customizeMetrics, alert)
	if err != nil {
		api.Errors.Internal(err)
	}

	if alert.AlertType == "" {
		alert.AlertType = "micro_service_customize"
	}
	alert.AlertScope = MicroServiceScope
	alert.AlertScopeID = params.TenantGroup
	alert.Attributes = nil
	for _, rule := range alert.Rules {
		rule.Attributes = map[string]interface{}{}
		rule.Attributes[Scope] = fmt.Sprintf("{{%s}}", tk)
		scopeFilter := adapt.CustomizeAlertRuleFilter{}
		scopeFilter.Tag = "_metric_scope"
		scopeFilter.Operator = OperateEq
		scopeFilter.Value = MicroServiceScope
		rule.Filters = append(rule.Filters, &scopeFilter)

		scopeIDFilter := adapt.CustomizeAlertRuleFilter{}
		scopeIDFilter.Tag = "_metric_scope"
		scopeIDFilter.Operator = OperateEq
		scopeIDFilter.Value = tk
		rule.Filters = append(rule.Filters, &scopeIDFilter)

		scopeApplicationFilter := adapt.CustomizeAlertRuleFilter{}
		scopeApplicationFilter.Tag = "application_id"
		scopeApplicationFilter.Operator = OperateIn
		scopeApplicationFilter.Value = "$" + "application_id"
		rule.Filters = append(rule.Filters, &scopeApplicationFilter)
	}

	alert.Lang = api.Language(r)
	err = p.microAlertAPI.CheckCustomizeAlert(&alert)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	id, err := p.microAlertAPI.CreateCustomizeAlert(&alert)
	if err != nil {
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists("alert")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"id": id,
	})
}

func (p *provider) updateCustomizeAlert(params struct {
	ID          uint64 `param:"id" validate:"required,gt=0"`
	TenantGroup string `query:"tenantGroup" validate:"required"`
}, alert adapt.CustomizeAlertDetail, r *http.Request) interface{} {
	customAlert, err := p.microAlertAPI.CustomizeAlertDetail(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	if customAlert.AlertScope != MicroServiceScope || customAlert.AlertScopeID != params.TenantGroup {
		return api.Errors.AccessDenied()
	}
	tk, err := p.authDb.InstanceTenant.QueryTkByTenantGroup(params.TenantGroup)
	if err != nil {
		return api.Errors.Internal("get tk failed.")
	}
	customizeMetrics, err := p.microAlertAPI.CustomizeMetrics(api.Language(r), MicroServiceScope, tk, nil)
	if err != nil {
		return api.Errors.Internal("get metric meta failed")
	}
	err = checkCustomMetric(customizeMetrics, alert)
	if err != nil {
		api.Errors.Internal(err)
	}

	alert.ID = params.ID
	alert.Enable = customAlert.Enable
	alert.AlertScope = MicroServiceScope
	alert.AlertScopeID = params.TenantGroup
	alert.Attributes = customAlert.Attributes
	for _, rule := range alert.Rules {
		rule.Attributes = map[string]interface{}{}
		rule.Attributes[Scope] = fmt.Sprintf("{{%s}}", tk)
		scopeFilter := adapt.CustomizeAlertRuleFilter{}
		scopeFilter.Tag = MetricScope
		scopeFilter.Operator = "eq"
		scopeFilter.Value = MicroServiceScope
		rule.Filters = append(rule.Filters, &scopeFilter)

		scopeIDFilter := adapt.CustomizeAlertRuleFilter{}
		scopeIDFilter.Tag = MetricScopeId
		scopeIDFilter.Operator = "eq"
		scopeIDFilter.Value = tk
		rule.Filters = append(rule.Filters, &scopeIDFilter)

		scopeApplicationFilter := adapt.CustomizeAlertRuleFilter{}
		scopeApplicationFilter.Tag = "application_id"
		scopeApplicationFilter.Operator = OperateIn
		scopeApplicationFilter.Value = "$" + "application_id"
		rule.Filters = append(rule.Filters, &scopeApplicationFilter)
	}

	err = p.microAlertAPI.CheckCustomizeAlert(&alert)

	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	alert.ID = params.ID
	err = p.microAlertAPI.UpdateCustomizeAlert(&alert)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) updateCustomizeAlertEnable(params struct {
	ID      uint64 `param:"id" validate:"required,gt=0"`
	Enable  bool   `param:"enable"`
	ScopeID string `query:"tenantGroup" validate:"required"`
}) interface{} {
	data, err := p.microAlertAPI.CustomizeAlert(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data == nil {
		return api.Errors.NotFound("id = " + strconv.FormatUint(params.ID, 10))
	}
	if data.AlertScopeID != params.ScopeID {
		return api.Errors.AccessDenied()
	}
	err = p.microAlertAPI.UpdateCustomizeAlertEnable(params.ID, params.Enable)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) deleteCustomizeAlert(params struct {
	ID      uint64 `param:"id" validate:"required,gt=0"`
	ScopeID string `query:"tenantGroup" validate:"required"`
}) interface{} {
	data, err := p.microAlertAPI.CustomizeAlert(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}

	if data == nil {
		return api.Errors.NotFound("id = " + strconv.FormatUint(params.ID, 10))
	}

	if data.AlertScopeID != params.ScopeID {
		return api.Errors.AccessDenied()
	}

	err = p.microAlertAPI.DeleteCustomizeAlert(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"name": data.Name,
	})
}
