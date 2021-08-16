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

package adapt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/monitor/utils"
)

// FromCustomizeAlertRule .
func (r *AlertRule) FromCustomizeAlertRule(lang i18n.LanguageCodes, t i18n.Translator, cr *db.CustomizeAlertRule) *AlertRule {
	rule, err := (&CustomizeAlertRule{}).FromModel(cr)
	if err != nil {
		return nil
	}
	r.ID = cr.ID
	r.Name = cr.Name
	r.AlertType = cr.AlertType
	r.AlertScope = cr.AlertScope
	r.AlertIndex = &DisplayKey{
		Key:     cr.AlertIndex,
		Display: t.Text(lang, cr.Name),
	}
	r.Template = cr.Template
	r.Window = int64(rule.Window)
	r.Attributes = cr.Attributes
	r.Enable = cr.Enable
	r.CreateTime = cr.CreateTime.UnixNano() / int64(time.Millisecond)
	r.UpdateTime = cr.UpdateTime.UnixNano() / int64(time.Millisecond)
	r.IsRecover = getRecoverFromAttributes(rule.Attributes)
	for _, item := range rule.Functions {
		function := &AlertRuleFunction{
			Field: &DisplayKey{
				Key:     item.Field,
				Display: item.Field,
			},
			Aggregator: item.Aggregator,
			Operator:   item.Operator,
			Value:      item.Value,
			DataType:   item.DataType,
			Unit:       item.Unit,
		}
		r.Functions = append(r.Functions, function)
	}
	return r
}

// FromModel .
func (r *AlertRule) FromModel(lang i18n.LanguageCodes, t i18n.Translator, m *db.AlertRule) *AlertRule {
	r.ID = m.ID
	r.Name = m.Name
	r.AlertType = m.AlertType
	r.AlertScope = m.AlertScope
	r.AlertIndex = &DisplayKey{
		Key:     m.AlertIndex,
		Display: t.Text(lang, m.Name), // default as name
	}
	r.Template = m.Template
	r.IsRecover = getRecoverFromAttributes(m.Attributes)
	r.Attributes = m.Attributes
	r.Version = m.Version
	r.Enable = m.Enable
	r.CreateTime = m.CreateTime.UnixNano() / int64(time.Millisecond)
	r.UpdateTime = m.UpdateTime.UnixNano() / int64(time.Millisecond)
	window, ok := utils.GetMapValueInt64(m.Template, "window")
	if !ok {
		return nil
	}
	r.Window = window
	functions, ok := utils.GetMapValueArr(m.Template, "functions")
	if !ok {
		return nil
	}
	for _, f := range functions {
		function, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		field, ok := utils.GetMapValueString(function, "field")
		if !ok {
			continue
		}
		aggregator, ok := utils.GetMapValueString(function, "aggregator")
		if !ok {
			continue
		}
		operator, ok := utils.GetMapValueString(function, "operator")
		if !ok {
			continue
		}
		value, ok := function["value"]
		if !ok {
			continue
		}
		unit, _ := utils.GetMapValueString(function, "unit")
		dataType := TypeOf(value)
		if dataType == "" {
			continue
		}
		r.Functions = append(r.Functions, &AlertRuleFunction{
			Field: &DisplayKey{
				Key:     field,
				Display: field,
			},
			Aggregator: aggregator,
			Operator:   operator,
			Value:      value,
			DataType:   dataType,
			Unit:       unit,
		})
	}
	return r
}

// FromModel .
func (n *AlertNotify) FromModel(m *db.AlertNotify, notifyGroupMap map[int64]*apistructs.NotifyGroup) *AlertNotify {
	t, unit := convertMillisecondToUnit(m.Silence)
	n.ID = m.ID
	n.Silence = &AlertNotifySilence{Value: t, Unit: unit, Policy: m.SilencePolicy}
	n.CreateTime = m.Created.UnixNano() / int64(time.Millisecond)
	n.UpdateTime = m.Updated.UnixNano() / int64(time.Millisecond)
	// fill in the alarm notification target
	notifyType, ok := utils.GetMapValueString(m.NotifyTarget, "type")
	if !ok {
		return nil
	}
	if notifyType == "notify_group" {
		// notify group ID
		groupID, ok := utils.GetMapValueInt64(m.NotifyTarget, "group_id")
		if !ok {
			return nil
		}
		// notify group method
		groupType, ok := utils.GetMapValueString(m.NotifyTarget, "group_type")
		if !ok {
			return nil
		}
		n.Type = notifyType
		n.GroupID = groupID
		n.GroupType = groupType
		n.NotifyGroup = notifyGroupMap[groupID]
	} else if notifyType == "dingding" {
		dingdingURL, ok := utils.GetMapValueString(m.NotifyTarget, "dingding_url")
		if !ok {
			return nil
		}
		n.Type = notifyType
		n.DingdingURL = dingdingURL
	} else {
		return nil
	}
	return n
}

// ToModel .
func (n *AlertNotify) ToModel(alert *Alert, silencePolicies map[string]bool) *db.AlertNotify {
	if alert == nil || n.Silence == nil {
		return nil
	}
	t := convertMillisecondByUnit(n.Silence.Value, n.Silence.Unit)
	if t < 0 {
		return nil
	}
	if n.Silence.Policy == "" || !silencePolicies[n.Silence.Policy] {
		n.Silence.Policy = fixedSliencePolicy
	}
	target := make(map[string]interface{})
	if n.GroupID != 0 && n.GroupType != "" {
		target["type"] = "notify_group"
		target["group_id"] = n.GroupID
		target["group_type"] = n.GroupType
	} else if n.DingdingURL != "" {
		target["type"] = "dingding"
		target["dingding_url"] = n.DingdingURL
	} else {
		return nil
	}
	return &db.AlertNotify{
		ID:             n.ID,
		AlertID:        alert.ID,
		NotifyTarget:   target,
		NotifyTargetID: "",
		Silence:        t,
		SilencePolicy:  n.Silence.Policy,
		Enable:         alert.Enable,
	}
}

// FromModel .
func (e *AlertExpression) FromModel(expression *db.AlertExpression) *AlertExpression {
	e.ID = expression.ID
	e.CreateTime = expression.Created.Unix()
	e.UpdateTime = expression.Updated.Unix()
	ruleIDStr, ok := utils.GetMapValueString(expression.Attributes, "rule_id")
	if !ok {
		return nil
	}
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 64)
	if err != nil {
		return nil
	}
	e.RuleID = ruleID
	alertIndex, ok := utils.GetMapValueString(expression.Attributes, "alert_index")
	if !ok {
		return nil
	}
	e.AlertIndex = alertIndex
	e.IsRecover = getRecoverFromAttributes(expression.Attributes)
	window, ok := utils.GetMapValueInt64(expression.Expression, "window")
	if !ok {
		return nil
	}
	e.Window = window
	functions, ok := utils.GetMapValueArr(expression.Expression, "functions")
	if !ok {
		return nil
	}
	for _, functionValue := range functions {
		function, ok := functionValue.(map[string]interface{})
		if !ok {
			continue
		}
		field, ok := utils.GetMapValueString(function, "field")
		if !ok {
			continue
		}
		aggregator, ok := utils.GetMapValueString(function, "aggregator")
		if !ok {
			continue
		}
		operator, ok := utils.GetMapValueString(function, "operator")
		if !ok {
			continue
		}
		value, ok := function["value"]
		if !ok {
			continue
		}
		e.Functions = append(e.Functions, &AlertExpressionFunction{
			Field:      field,
			Aggregator: aggregator,
			Operator:   operator,
			Value:      value,
		})
	}
	return e
}

// ToModel .
func (e *AlertExpression) ToModel(orgName string, alert *Alert, rule *AlertRule) (*db.AlertExpression, error) {
	attributes := make(map[string]interface{})
	//org, err := new(bundle.Bundle).GetOrg(alert.AlertScopeID)
	//if err != nil {
	//	return nil, err
	//}
	//attributes["org_name"] = org.Name
	for k, v := range rule.Attributes {
		attributes[k] = v
	}
	for k, v := range alert.Attributes {
		attributes[k] = v
	}
	attributes["alert_id"] = strconv.FormatUint(alert.ID, 10)
	attributes["rule_id"] = strconv.FormatUint(rule.ID, 10)
	attributes["alert_type"] = rule.AlertType
	attributes["alert_index"] = rule.AlertIndex.Key
	attributes["alert_name"] = rule.Name
	attributes["alert_title"] = alert.Name
	attributes["alert_scope"] = alert.AlertScope
	attributes["alert_scope_id"] = alert.AlertScopeID
	attributes["recover"] = strconv.FormatBool(e.IsRecover)
	// remove some fields that are not needed by flink to avoid too long attributes
	for _, item := range []string{"alert_domain", "alert_dashboard_id", "alert_dashboard_path", "alert_record_path"} {
		delete(attributes, item)
	}

	expression := rule.Template
	expression["window"] = e.Window
	// fill expression filters
	filters, _ := utils.GetMapValueArr(expression, "filters")
	for _, filterValue := range filters {
		filterMap, ok := filterValue.(map[string]interface{})
		if !ok {
			continue
		}
		tag, ok := utils.GetMapValueString(filterMap, "tag")
		if !ok {
			continue
		}
		operator, ok := utils.GetMapValueString(filterMap, "operator")
		if !ok {
			continue
		}
		opType := filterOperatorRel[operator]
		value, ok := filterMap["value"]
		if !ok {
			continue
		}
		if tag == clusterNameTag || tag == applicationIdTag {
			v, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("assert cluster_name or application_id is failed")
			}
			if !strings.HasPrefix(v, "$") {
				continue
			}
		}
		if attr, ok := attributes[tag]; ok {
			val, err := formatOperatorValue(opType, utils.StringType, attr)
			if err != nil {
				return nil, err
			}
			value = val
		}
		filterMap["value"] = value
	}
	expression["filters"] = filters

	// fill expression functions
	functionMap := make(map[string]*AlertExpressionFunction)
	for _, function := range e.Functions {
		functionMap[function.Aggregator+"-"+function.Field] = function
	}
	functions, _ := utils.GetMapValueArr(expression, "functions")
	for _, functionValue := range functions {
		function, ok := functionValue.(map[string]interface{})
		if !ok {
			continue
		}
		field, ok := utils.GetMapValueString(function, "field")
		if !ok {
			continue
		}
		aggregator, _ := utils.GetMapValueString(function, "aggregator")
		if !ok {
			continue
		}
		newFunction, ok := functionMap[aggregator+"-"+field]
		if !ok {
			continue
		}
		operator, _ := utils.GetMapValueString(function, "operator")
		if operator != "" && newFunction.Operator != "" {
			operator = newFunction.Operator
		}
		function["operator"] = operator

		value, _ := function["value"]
		if value != nil && newFunction.Value != nil {
			opType := functionOperatorRel[operator]
			dataType := utils.TypeOf(value)
			val, err := formatOperatorValue(opType, dataType, newFunction.Value)
			if err != nil {
				return nil, err
			}
			function["value"] = val
		}
	}
	expression["functions"] = functions

	// transform alert url
	alertDomain := alert.Domain
	if alertDomain == "" {
		if s, ok := utils.GetMapValueString(attributes, "alert_domain"); ok {
			alertDomain = s
		}
	}
	if strings.HasSuffix(alertDomain, "/") {
		alertDomain = alertDomain[0 : len(alertDomain)-1]
	}
	if routeID, ok := utils.GetMapValueString(rule.Attributes, "display_url_id"); ok {
		if alertURL := convertAlertURL(alertDomain, orgName, routeID, attributes); alertURL != "" {
			attributes["display_url"] = alertURL
		}
	}
	if dashboardID, ok := utils.GetMapValueString(rule.Attributes, "alert_dashboard_id"); ok {
		if dashboardPath, ok := utils.GetMapValueString(alert.Attributes, "alert_dashboard_path"); ok {
			// get group's value
			var groups []string
			group, _ := utils.GetMapValueArr(expression, "group")
			for _, item := range group {
				groups = append(groups, item.(string))
			}
			attributes["display_url"] = convertDashboardURL(alertDomain, orgName, dashboardPath, dashboardID, groups)
		}
	}

	// transform record url
	if recordPath, ok := utils.GetMapValueString(alert.Attributes, "alert_record_path"); ok {
		attributes["record_url"] = convertRecordURL(alertDomain, orgName, recordPath)
	}

	return &db.AlertExpression{
		ID:         e.ID,
		AlertID:    alert.ID,
		Expression: expression,
		Attributes: attributes,
		Version:    "3.0",
		Enable:     alert.Enable,
	}, nil
}

// format operation
func formatOperatorValue(opType string, dataType string, value interface{}) (interface{}, error) {
	if opType == "" || dataType == "" {
		return nil, invalidParameter("%s not support %s data type", opType, dataType)
	}
	switch opType {
	case OperatorTypeNone:
		return nil, nil
	case OperatorTypeOne:
		val, err := convertDataByType(value, dataType)
		if err != nil {
			return nil, err
		}
		return val, nil
	case OperatorTypeMore:
		switch value.(type) {
		case string, []byte:
			value, ok := convertString(value)
			if !ok {
				return nil, invalidParameter("%s not support %v data", opType, value)
			}
			var values []interface{}
			for _, val := range strings.Split(value, ",") {
				v, err := convertDataByType(val, dataType)
				if err != nil {
					return nil, err
				}
				values = append(values, v)
			}
			return values, nil
		default:
			var values []interface{}
			valueOf := reflect.ValueOf(value)
			switch valueOf.Kind() {
			case reflect.Slice, reflect.Array:
				for i := 0; i < valueOf.Len(); i++ {
					val, err := convertDataByType(valueOf.Index(i).Interface(), dataType)
					if err != nil {
						return nil, err
					}
					values = append(values, val)
				}
			default:
				val, err := convertDataByType(value, dataType)
				if err != nil {
					return nil, err
				}
				values = append(values, val)
			}
			return values, nil
		}
	}
	return nil, invalidParameter("%s not support", opType)
}

// FromModel .
func (a *Alert) FromModel(m *db.Alert) *Alert {
	a.ID = m.ID
	a.Name = m.Name
	a.AlertScope = m.AlertScope
	a.AlertScopeID = m.AlertScopeID
	a.Attributes = m.Attributes
	a.Enable = m.Enable
	a.CreateTime = m.Created.UnixNano() / int64(time.Millisecond)
	a.UpdateTime = m.Updated.UnixNano() / int64(time.Millisecond)
	return a
}

// ToModel .
func (a *Alert) ToModel() *db.Alert {
	return &db.Alert{
		ID:           a.ID,
		Name:         a.Name,
		AlertScope:   a.AlertScope,
		AlertScopeID: a.AlertScopeID,
		Attributes:   a.Attributes,
		Enable:       a.Enable,
	}
}

func getRecoverFromAttributes(m map[string]interface{}) bool {
	if s, ok := utils.GetMapValueString(m, "recover"); ok {
		if v, err := strconv.ParseBool(s); err == nil {
			return v
		}
	}
	return false
}
