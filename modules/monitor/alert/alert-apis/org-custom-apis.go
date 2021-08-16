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

package apis

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"github.com/erda-project/erda/modules/monitor/common/permission"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
	"github.com/erda-project/erda/modules/monitor/utils"
)

func (p *provider) queryOrgCustomizeMetric(r *http.Request, params struct {
}) interface{} {
	orgID := api.OrgID(r)
	org, err := p.bdl.GetOrg(orgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	lang := api.Language(r)
	cms, err := p.a.CustomizeMetrics(lang, "org", org.Name, nil)
	if err != nil {
		return api.Errors.Internal(err)
	}

	for _, metric := range cms.Metrics {
		tags := make([]*adapt.TagMeta, 0)
		for _, tag := range metric.Tags {
			if p.orgFilterTags[tag.Tag.Key] {
				continue
			}
			tags = append(tags, tag)
		}
		metric.Tags = tags
	}
	var functionOperators []*adapt.Operator
	for _, op := range cms.FunctionOperators {
		if op.Type == adapt.OperatorTypeOne {
			functionOperators = append(functionOperators, op)
		}
	}
	cms.FunctionOperators = functionOperators
	var filterOperators []*adapt.Operator
	for _, op := range cms.FilterOperators {
		if op.Type == adapt.OperatorTypeOne {
			filterOperators = append(filterOperators, op)
		}
	}
	cms.FilterOperators = filterOperators
	if lang == nil {
		cms.NotifySample = adapt.OrgNotifyTemplateSample
	} else {
		cms.NotifySample = adapt.OrgNotifyTemplateSampleEn
		for _, v := range lang {
			if strings.HasPrefix(v.Code, "zh") {
				cms.NotifySample = adapt.OrgNotifyTemplateSample
			}
		}
	}
	return api.Success(cms)
}

func (p *provider) queryOrgCustomizeAlerts(r *http.Request, params struct {
	PageNo   int `query:"pageNo" validate:"gte=1"`
	PageSize int `query:"pageSize" validate:"gte=1,lte=100"`
}) interface{} {
	orgID := api.OrgID(r)
	alert, total, err := p.a.CustomizeAlerts(api.Language(r), "org", orgID, params.PageNo, params.PageSize)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(map[string]interface{}{
		"total": total,
		"list":  alert,
	})
}

func (p *provider) getOrgCustomizeAlertDetail(ctx httpserver.Context, params struct {
	ID uint64 `param:"id" validate:"required,gt=0"`
}) interface{} {
	orgID := api.OrgID(ctx.Request())
	alert, err := p.a.CustomizeAlertDetail(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	if alert.AlertScope != "org" && alert.AlertScopeID != orgID {
		permission.Failure(ctx, nil)
		return nil
	}
	return api.Success(alert)
}

func (p *provider) createOrgCustomizeAlert(r *http.Request, alert *adapt.CustomizeAlertDetail) interface{} {
	orgID := api.OrgID(r)
	if alert.AlertType == "" {
		alert.AlertType = "org_customize"
	}
	org, err := p.bdl.GetOrg(orgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	alert.Lang = api.Language(r)
	alert.AlertScope = "org"
	alert.AlertScopeID = orgID
	alert.Attributes = make(map[string]interface{})

	// Get all metrics. Currently there is only one metricName
	var metricNames []string
	for _, rule := range alert.Rules {
		metricNames = append(metricNames, rule.Metric)
	}
	lang := i18n.LanguageCodes{
		{
			Code: "",
		},
	}
	metricMeta, err := p.metricq.MetricMeta(lang, alert.AlertScope, org.Name, metricNames...)
	if err != nil {
		return api.Errors.Internal(err)
	}
	metricMap := make(map[string]*metrics.MetricMeta)
	for _, metric := range metricMeta {
		metricMap[metric.Name.Key] = metric
	}

	if len(metricNames) <= 0 {
		return api.Errors.Internal(err)
	}

	for _, rule := range alert.Rules {
		rule.Attributes = make(map[string]interface{})
		rule.Attributes["alert_group"] = "{{cluster_name}}"

		ruleMetric := metricMap[rule.Metric]
		labels := ruleMetric.Labels
		scope := labels["metric_scope"]
		scopeId := org.Name

		if err := p.checkMetricMeta(rule, metricMap[rule.Metric]); err != nil {
			return api.Errors.InvalidParameter(err)
		}

		if scope != "" {
			rule.Filters = append(rule.Filters, &adapt.CustomizeAlertRuleFilter{
				Tag:      "_metric_scope",
				Operator: "eq",
				Value:    scope,
			})
		}
		if scopeId != "" {
			rule.Filters = append(rule.Filters, &adapt.CustomizeAlertRuleFilter{
				Tag:      "_metric_scope_id",
				Operator: "eq",
				Value:    scopeId,
			})
		}
		rule.Filters = append(rule.Filters, &adapt.CustomizeAlertRuleFilter{
			Tag:      "cluster_name",
			Operator: "in",
			Value:    "$cluster_name",
		})
	}
	if err := p.checkCustomizeAlert(alert); err != nil {
		return api.Errors.InvalidParameter(err)
	}

	id, err := p.a.CreateCustomizeAlert(alert)
	if err != nil {
		if adapt.IsAlreadyExistsError(err) {
			return api.Errors.AlreadyExists("alert")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(id)
}

func (p *provider) updateOrgCustomizeAlert(r *http.Request, ctx httpserver.Context, newAlert *adapt.CustomizeAlertDetail) interface{} {
	orgID := api.OrgID(r)
	if newAlert.AlertType == "" {
		newAlert.AlertType = "org_customize"
	}
	org, err := p.bdl.GetOrg(orgID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	alert, err := p.a.CustomizeAlert(newAlert.ID)
	if err != nil {
		return api.Errors.Internal(err)
	} else if alert == nil {
		return api.Errors.NotFound(err)
	}
	if alert.AlertScope != "org" && alert.AlertScopeID != orgID {
		permission.Failure(ctx, nil)
		return nil
	}

	newAlert.AlertScope = alert.AlertScope
	newAlert.AlertScopeID = alert.AlertScopeID
	newAlert.Enable = alert.Enable
	newAlert.ID = alert.ID

	// get metric
	var metricNames []string
	for _, rule := range newAlert.Rules {
		metricNames = append(metricNames, rule.Metric)
	}

	lang := i18n.LanguageCodes{
		{
			Code: "",
		},
	}

	metricMeta, err := p.metricq.MetricMeta(lang, alert.AlertScope, org.Name, metricNames...)
	if err != nil {
		return api.Errors.Internal(err)
	}
	metricMap := make(map[string]*metrics.MetricMeta)
	for _, metric := range metricMeta {
		metricMap[metric.Name.Key] = metric
	}

	for _, rule := range newAlert.Rules {
		metric := metricMap[rule.Metric]
		if err := p.checkMetricMeta(rule, metric); err != nil {
			return api.Errors.InvalidParameter(err)
		}
		if len(metric.Name.Name) > 0 {
			if rule.Attributes == nil {
				rule.Attributes = make(map[string]interface{})
			}
			rule.Attributes["metric_name"] = metric.Name.Name
		}
	}

	if err := p.checkCustomizeAlert(newAlert); err != nil {
		return api.Errors.InvalidParameter(err)
	}

	err = p.a.UpdateCustomizeAlert(newAlert)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(true)
}

func (p *provider) checkMetricMeta(
	rule *adapt.CustomizeAlertRule, metric *metrics.MetricMeta) error {
	if metric == nil {
		return fmt.Errorf("rule metric is not match")
	} else if len(rule.Functions) == 0 {
		return fmt.Errorf("rule functions is empty")
	}

	selects := make(map[string]string)
	tagRel := make(map[string]bool)
	for _, tag := range metric.Tags {
		tagRel[tag.Key] = true
		if ok := p.orgFilterTags[tag.Key]; !ok || tag.Key == "cluster_name" {
			selects[tag.Key] = "#" + tag.Key
		}
	}

	fieldRel := make(map[string]string)
	for _, field := range metric.Fields {
		fieldRel[field.Key] = field.Type
	}

	aliasRel := make(map[string]bool)
	// Check calculation function
	for _, function := range rule.Functions {
		if function.Alias == "" {
			return fmt.Errorf("function %s alias can not be empty", function.Field)
		}
		if aliasRel[function.Alias] {
			return fmt.Errorf("alias %s duplicate", function.Alias)
		}
		aliasRel[function.Alias] = true
		dataType, field := "string", function.Field
		if strings.HasPrefix(function.Field, "fields.") {
			field = function.Field[len("fields."):]
		}
		dt, ok := fieldRel[field]
		if ok {
			dataType = dt
		}
		opType, ok := p.a.FunctionOperatorKeysMap()[function.Operator]
		if !ok {
			return fmt.Errorf("not support rule function operator %s", function.Operator)
		}
		if _, ok := p.a.AggregatorKeysSet()[function.Aggregator]; !ok {
			return fmt.Errorf("not support rule function aggregator %s", function.Aggregator)
		}

		// According to the data type and operation type conversion threshold
		value, apiErr := p.formatOperatorValue(opType, dataType, function.Value)
		if apiErr != nil {
			return apiErr
		}
		function.Value = value
	}

	for _, filter := range rule.Filters {
		if ok := tagRel[filter.Tag]; !ok {
			return fmt.Errorf(fmt.Sprintf("not support rule filter tag %s", filter.Tag))
		}
		opType, ok := p.a.FilterOperatorKeysMap()[filter.Operator]
		if !ok {
			return fmt.Errorf(fmt.Sprintf("not support rule filter operator %s", filter.Operator))
		}
		if utils.StringType != utils.TypeOf(filter.Value) {
			return fmt.Errorf(fmt.Sprintf("not support rule filter value %v", filter.Value))
		}

		// 根据数据类型和操作类型转换阈值
		value, apiErr := p.formatOperatorValue(opType, utils.StringType, filter.Value)
		if apiErr != nil {
			return apiErr
		}
		filter.Value = value
	}

	for _, group := range rule.Group {
		if _, ok := tagRel[group]; !ok {
			return fmt.Errorf(fmt.Sprintf("not support rule group tag %s", group))
		}
	}

	rule.Outputs = []string{"alert"}
	rule.Select = selects
	return nil
}

func (p *provider) formatOperatorValue(
	opType string, dataType string, value interface{}) (interface{}, error) {
	if opType == "" || dataType == "" {
		return nil, fmt.Errorf(fmt.Sprintf("%s not support %s data type", opType, dataType))
	}

	switch opType {
	case "none":
		return nil, nil
	case "one":
		if val, err := utils.ConvertDataType(value, dataType); err != nil {
			return nil, err
		} else {
			return val, nil
		}
	case "more":
		switch value.(type) {
		case string, []byte:
			value, ok := utils.ConvertString(value)
			if !ok {
				return nil, fmt.Errorf(fmt.Sprintf("%s not support %v data", opType, value))
			}
			var values []interface{}
			for _, val := range strings.Split(value, ",") {
				if v, err := utils.ConvertDataType(val, dataType); err != nil {
					return nil, err
				} else {
					values = append(values, v)
				}
			}
			return values, nil
		default:
			var values []interface{}
			valueOf := reflect.ValueOf(value)
			switch valueOf.Kind() {
			case reflect.Slice, reflect.Array:
				for i := 0; i < valueOf.Len(); i++ {
					if val, err := utils.ConvertDataType(valueOf.Index(i).Interface(), dataType); err != nil {
						return nil, err
					} else {
						values = append(values, val)
					}
				}
			default:
				if val, err := utils.ConvertDataType(value, dataType); err != nil {
					return nil, err
				} else {
					values = append(values, val)
				}
			}
			return values, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("%s not support", opType))
}

func (p *provider) updateOrgCustomizeAlertEnable(params struct {
	ID     int  `param:"id" validate:"required,gt=0"`
	Enable bool `param:"enable"`
}) interface{} {
	err := p.a.UpdateCustomizeAlertEnable(uint64(params.ID), params.Enable)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(nil)
}

func (p *provider) deleteOrgCustomizeAlert(ctx httpserver.Context, params struct {
	ID uint64 `param:"id" validate:"required,gt=0"`
}) interface{} {
	data, _ := p.a.CustomizeAlert(params.ID)
	err := p.a.DeleteCustomizeAlert(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	if data != nil {
		return api.Success(map[string]interface{}{
			"name": data.Name,
		})
	}
	return api.Success(true)
}
