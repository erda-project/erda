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
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type alertService struct {
	p *provider
}

func (m *alertService) QueryOrgDashboardByAlert(ctx context.Context, request *pb.QueryOrgDashboardByAlertRequest) (*pb.QueryOrgDashboardByAlertResponse, error) {
	orgID := apis.GetOrgID(ctx)
	if request.AlertType == "" {
		request.AlertType = "org_customize"
	}
	_, err := m.p.bdl.GetOrg(orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	request.AlertScope = "org"
	request.AlertScopeId = orgID
	if request.AlertScopeId == "" {
		return nil, fmt.Errorf("alert scope id must not be empty")
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	detail := &pb.CustomizeAlertDetail{}
	err = json.Unmarshal(data, detail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	dash, err := adapt.NewDashboard(m.p.a).GenerateDashboardPreView(detail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.QueryOrgDashboardByAlertResponse{
		Data: dash,
	}
	return result, nil
}

func (m *alertService) CreateAlert(ctx context.Context, request *pb.CreateAlertRequest) (*pb.CreateAlertResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert := &pb.Alert{}
	err = json.Unmarshal(data, alert)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if response := m.p.checkAlert(alert); response != nil {
		res, ok := response.(*api.Response)
		if !ok {
			return nil, fmt.Errorf("res is not api.Response")
		}
		return nil, res.Err
	}
	orgID := alert.Attributes["dice_org_id"]
	org, err := m.p.bdl.GetOrg(orgID.AsInterface())
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert.Attributes["org_name"] = structpb.NewStringValue(org.Name)
	id, err := m.p.a.CreateAlert(alert)
	result := &pb.CreateAlertResponse{
		Data: id,
	}
	return result, nil
}

func (m *alertService) CreateOrgAlert(ctx context.Context, request *pb.CreateOrgAlertRequest) (*pb.CreateOrgAlertResponse, error) {
	orgID := apis.GetOrgID(ctx)
	org, err := m.p.bdl.GetOrg(orgID)
	if err != nil {
		return nil, errors.NewInvalidParameterError("orgId", "orgId is invalidate")
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert := &pb.Alert{}
	err = json.Unmarshal(data, alert)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert.Attributes = make(map[string]*structpb.Value)
	alert.Attributes["org_name"] = structpb.NewStringValue(org.Name)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("orgId", "orgId is invalidate")
	}
	if len(alert.ClusterNames) <= 0 {
		return nil, errors.NewMissingParameterError("cluster name")
	}
	if !m.checkOrgClusterNames(id, alert.ClusterNames) {
		return nil, errors.NewPermissionError("monitor_org_alert", "create", "access denied")
	}
	aid, err := m.p.a.CreateOrgAlert(alert, orgID)
	if err != nil {
		if adapt.IsInvalidParameterError(err) {
			return nil, errors.NewInvalidParameterError("alertScope", alert.AlertScope)
		}
		if adapt.IsAlreadyExistsError(err) {
			return nil, errors.NewAlreadyExistsError("monitor_org_alert")
		}
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.CreateOrgAlertResponse{
		Id: aid,
	}, nil
}

func (m *alertService) QueryCustomizeMetric(ctx context.Context, request *pb.QueryCustomizeMetricRequest) (*pb.QueryCustomizeMetricResponse, error) {
	result := &pb.QueryCustomizeMetricResponse{}
	lang := apis.Language(ctx)
	data, err := m.p.a.CustomizeMetrics(lang, request.Scope, request.ScopeId, request.Name)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result.Data = data
	return result, nil
}

func (m *alertService) QueryCustomizeNotifyTarget(ctx context.Context, request *pb.QueryCustomizeNotifyTargetRequest) (*pb.QueryCustomizeNotifyTargetResponse, error) {
	result := &pb.QueryCustomizeNotifyTargetResponse{
		Data: &pb.QueryCustomizeNotifyTargetData{},
	}
	lang := apis.Language(ctx)
	data := m.p.a.NotifyTargetsKeys(lang, apis.GetOrgID(ctx))
	result.Data.Targets = data
	return result, nil
}

func (m *alertService) QueryOrgCustomizeNotifyTarget(ctx context.Context, request *pb.QueryOrgCustomizeNotifyTargetRequest) (*pb.QueryOrgCustomizeNotifyTargetResponse, error) {
	result := &pb.QueryOrgCustomizeNotifyTargetResponse{
		Data: &pb.QueryCustomizeNotifyTargetData{},
	}
	lang := apis.Language(ctx)
	orgID := apis.GetOrgID(ctx)
	data := m.p.a.NotifyTargetsKeys(lang, orgID)
	result.Data.Targets = data
	return result, nil
}

func (m *alertService) QueryCustomizeAlert(ctx context.Context, request *pb.QueryCustomizeAlertRequest) (*pb.QueryCustomizeAlertResponse, error) {
	result := &pb.QueryCustomizeAlertResponse{
		Data: &pb.QueryCustomizeAlertData{
			List: make([]*pb.CustomizeAlertOverview, 0),
		},
	}
	lang := apis.Language(ctx)
	alert, total, err := m.p.a.CustomizeAlerts(lang, request.Scope, request.ScopeId, int(request.PageNo), int(request.PageSize))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result.Data.List = alert
	result.Data.Total = int64(total)
	return result, nil
}

func (m *alertService) GetCustomizeAlert(ctx context.Context, request *pb.GetCustomizeAlertRequest) (*pb.GetCustomizeAlertResponse, error) {
	alert, err := m.p.a.CustomizeAlert(request.Id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.GetCustomizeAlertResponse{
		Data: &pb.CustomizeAlertDetail{},
	}
	result.Data = alert
	return result, nil
}

func (m alertService) GetCustomizeAlertDetail(ctx context.Context, request *pb.GetCustomizeAlertDetailRequest) (*pb.GetCustomizeAlertDetailResponse, error) {
	alert, err := m.p.a.CustomizeAlertDetail(request.Id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.GetCustomizeAlertDetailResponse{
		Data: alert,
	}
	return result, nil
}

func (m *alertService) CreateCustomizeAlert(ctx context.Context, request *pb.CreateCustomizeAlertRequest) (*pb.CreateCustomizeAlertResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert := &pb.CustomizeAlertDetail{}
	err = json.Unmarshal(data, alert)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert.Rules = request.Rules
	alert.Notifies = request.Notifies
	alert.Attributes = request.Attributes
	err = m.checkCustomizeAlert(alert)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	id, err := m.p.a.CreateCustomizeAlert(alert)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	reslut := &pb.CreateCustomizeAlertResponse{
		Data: id,
	}
	return reslut, nil
}

func (m *alertService) checkCustomizeAlert(alert *pb.CustomizeAlertDetail) error {
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

func (m *alertService) UpdateCustomizeAlert(ctx context.Context, request *pb.UpdateCustomizeAlertRequest) (*pb.UpdateCustomizeAlertResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req := &pb.CustomizeAlertDetail{}
	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = m.p.checkCustomizeAlert(req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = m.p.a.UpdateCustomizeAlert(req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (m *alertService) UpdateCustomizeAlertEnable(ctx context.Context, request *pb.UpdateCustomizeAlertEnableRequest) (*pb.UpdateCustomizeAlertEnableResponse, error) {
	err := m.p.a.UpdateCustomizeAlertEnable(request.Id, request.Enable)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (m *alertService) DeleteCustomizeAlert(ctx context.Context, request *pb.DeleteCustomizeAlertRequest) (*pb.DeleteCustomizeAlertResponse, error) {
	data, _ := m.p.a.CustomizeAlert(request.Id)
	result := &pb.DeleteCustomizeAlertResponse{}
	err := m.p.a.DeleteCustomizeAlert(request.Id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if data != nil {
		result.Data = data.Name
		return result, nil
	}
	return result, nil
}

func (m *alertService) QueryOrgCustomizeMetric(ctx context.Context, request *pb.QueryOrgCustomizeMetricRequest) (*pb.QueryOrgCustomizeMetricResponse, error) {
	orgID := apis.GetOrgID(ctx)
	org, err := m.p.bdl.GetOrg(orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	lang := apis.Language(ctx)
	cms, err := m.p.a.CustomizeMetrics(lang, "org", org.Name, nil)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	for _, metric := range cms.Metrics {
		tags := make([]*pb.TagMeta, 0)
		for _, tag := range metric.Tags {
			if m.p.orgFilterTags[tag.Tag.Key] {
				continue
			}
			tags = append(tags, tag)
		}
		metric.Tags = tags
	}
	var functionOperators []*pb.Operator
	for _, op := range cms.FunctionOperators {
		if op.Type == adapt.OperatorTypeOne {
			functionOperators = append(functionOperators, op)
		}
	}
	cms.FunctionOperators = functionOperators
	var filterOperators []*pb.Operator
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
	result := &pb.QueryOrgCustomizeMetricResponse{
		Data: cms,
	}
	return result, nil
}

func (m *alertService) QueryOrgCustomizeAlerts(ctx context.Context, request *pb.QueryOrgCustomizeAlertsRequest) (*pb.QueryOrgCustomizeAlertsResponse, error) {
	orgID := apis.GetOrgID(ctx)
	language := apis.Language(ctx)
	alert, total, err := m.p.a.CustomizeAlerts(language, "org", orgID, int(request.PageNo), int(request.PageSize))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data := &pb.QueryOrgCustomizeAlertsResponse{
		Data: &pb.QueryOrgCustomizeAlertsData{
			List: make([]*pb.CustomizeAlertOverview, 0),
		},
	}
	data.Data.Total = int64(total)
	data.Data.List = alert
	return data, nil
}

func (m *alertService) GetOrgCustomizeAlertDetail(ctx context.Context, request *pb.GetOrgCustomizeAlertDetailRequest) (*pb.GetOrgCustomizeAlertDetailResponse, error) {
	orgID := apis.GetOrgID(ctx)
	alert, err := m.p.a.CustomizeAlertDetail(request.Id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if alert == nil {
		return nil, errors.NewNotFoundError("monitor_org_alert")
	}
	if alert.AlertScope != "org" && alert.AlertScopeId != orgID {
		return nil, errors.NewPermissionError("monitor_org_alert", "list", "access denied")
	}
	result := &pb.GetOrgCustomizeAlertDetailResponse{
		Data: &pb.CustomizeAlertDetail{},
	}
	result.Data = alert
	return result, nil
}

func (m *alertService) CreateOrgCustomizeAlert(ctx context.Context, request *pb.CreateOrgCustomizeAlertRequest) (*pb.CreateOrgCustomizeAlertResponse, error) {
	orgID := apis.GetOrgID(ctx)
	if request.AlertType == "" {
		request.AlertType = "org_customize"
	}
	org, err := m.p.bdl.GetOrg(orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	request.AlertScope = "org"
	request.AlertScopeId = orgID
	request.Attributes = make(map[string]*structpb.Value)
	var metricNames []string
	for _, rule := range request.Rules {
		metricNames = append(metricNames, rule.Metric)
	}
	lang := apis.Language(ctx)
	metricMeta, err := m.p.metricq.MetricMeta(lang, request.AlertScope, org.Name, metricNames...)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	metricMap := make(map[string]*metricpb.MetricMeta)
	for _, metric := range metricMeta {
		metricMap[metric.Name.Key] = metric
	}
	if len(metricNames) <= 0 {
		return nil, errors.NewInternalServerError(err)
	}
	for _, rule := range request.Rules {
		rule.Attributes = make(map[string]*structpb.Value)
		rule.Attributes["alert_group"] = structpb.NewStringValue("{{cluster_name}}")

		ruleMetric := metricMap[rule.Metric]
		labels := ruleMetric.Labels
		scope := labels["metric_scope"]
		scopeId := labels["metric_scope_id"]

		if err := m.checkMetricMeta(rule, metricMap[rule.Metric]); err != nil {
			return nil, errors.NewInternalServerError(err)
		}

		if scope != "" {
			rule.Filters = append(rule.Filters, &pb.CustomizeAlertRuleFilter{
				Tag:      "_metric_scope",
				Operator: "eq",
				Value:    structpb.NewStringValue(scope),
			})
		}
		if scopeId != "" {
			rule.Filters = append(rule.Filters, &pb.CustomizeAlertRuleFilter{
				Tag:      "_metric_scope_id",
				Operator: "eq",
				Value:    structpb.NewStringValue(scopeId),
			})
		}
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertDetail := &pb.CustomizeAlertDetail{}
	err = json.Unmarshal(data, alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if err := m.checkCustomizeAlert(alertDetail); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	id, err := m.p.a.CreateCustomizeAlert(alertDetail)
	if err != nil {
		if adapt.IsAlreadyExistsError(err) {
			return nil, errors.NewAlreadyExistsError("alert")
		}
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.CreateOrgCustomizeAlertResponse{
		Id: id,
	}
	return result, nil
}

func (m *alertService) checkMetricMeta(
	rule *pb.CustomizeAlertRule, metric *metricpb.MetricMeta) error {
	if metric == nil {
		return fmt.Errorf("rule metric is not match")
	} else if len(rule.Functions) == 0 {
		return fmt.Errorf("rule functions is empty")
	}

	selects := make(map[string]string)
	tagRel := make(map[string]bool)
	for _, tag := range metric.Tags {
		tagRel[tag.Key] = true
		if ok := m.p.orgFilterTags[tag.Key]; !ok || tag.Key == "cluster_name" {
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
		opType, ok := m.p.a.FunctionOperatorKeysMap()[function.Operator]
		if !ok {
			return fmt.Errorf("not support rule function operator %s", function.Operator)
		}
		if _, ok := m.p.a.AggregatorKeysSet()[function.Aggregator]; !ok {
			return fmt.Errorf("not support rule function aggregator %s", function.Aggregator)
		}

		// According to the data type and operation type conversion threshold
		value, apiErr := m.formatOperatorValue(opType, dataType, function.Value.AsInterface())
		if apiErr != nil {
			return apiErr
		}
		valueData, err := structpb.NewValue(value)
		if err != nil {
			logrus.Errorf("value transform is fail value is %s", value)
		}
		function.Value = valueData
	}

	for _, filter := range rule.Filters {
		if ok := tagRel[filter.Tag]; !ok {
			return fmt.Errorf(fmt.Sprintf("not support rule filter tag %s", filter.Tag))
		}
		opType, ok := m.p.a.FilterOperatorKeysMap()[filter.Operator]
		if !ok {
			return fmt.Errorf(fmt.Sprintf("not support rule filter operator %s", filter.Operator))
		}
		filterValue := filter.Value.AsInterface()
		if utils.StringType != utils.TypeOf(filterValue) {
			return fmt.Errorf(fmt.Sprintf("not support rule filter value %v", filter.Value))
		}

		// 根据数据类型和操作类型转换阈值
		value, apiErr := m.formatOperatorValue(opType, utils.StringType, filter.Value)
		if apiErr != nil {
			return apiErr
		}
		valueData, err := structpb.NewValue(value)
		if err != nil {
			logrus.Errorf("value is transform is fail value is %v", value)
			return err
		}
		filter.Value = valueData
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

func (m *alertService) formatOperatorValue(
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

func (m *alertService) UpdateOrgCustomizeAlert(ctx context.Context, request *pb.UpdateOrgCustomizeAlertRequest) (*pb.UpdateOrgCustomizeAlertResponse, error) {
	orgID := apis.GetOrgID(ctx)
	if request.AlertType == "" {
		request.AlertType = "org_customize"
	}
	org, err := m.p.bdl.GetOrg(orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert, err := m.p.a.CustomizeAlert(request.Id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	} else if alert == nil {
		return nil, errors.NewNotFoundError(request.Name)
	}
	if request.AlertScope != "org" && request.AlertScopeId != orgID {
		return nil, errors.NewPermissionError("monitor_org_alert", "update", "scope or scopeId is invalidate")
	}
	request.AlertScope = alert.AlertScope
	request.AlertScopeId = alert.AlertScopeId
	request.Enable = alert.Enable
	var metricNames []string
	for _, rule := range request.Rules {
		metricNames = append(metricNames, rule.Metric)
	}
	lang := apis.Language(ctx)
	metricMeta, err := m.p.metricq.MetricMeta(lang, alert.AlertScope, org.Name, metricNames...)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	metricMap := make(map[string]*metricpb.MetricMeta)
	for _, metric := range metricMeta {
		metricMap[metric.Name.Key] = metric
	}

	for _, rule := range request.Rules {
		metric := metricMap[rule.Metric]
		if err := m.checkMetricMeta(rule, metric); err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		if len(metric.Name.Name) > 0 {
			if rule.Attributes == nil {
				rule.Attributes = make(map[string]*structpb.Value)
			}
			rule.Attributes["metric_name"] = structpb.NewStringValue(metric.Name.Name)
		}
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertDetail := &pb.CustomizeAlertDetail{}
	err = json.Unmarshal(data, alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if err := m.checkCustomizeAlert(alertDetail); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = m.p.a.UpdateCustomizeAlert(alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.UpdateOrgCustomizeAlertResponse{
		Data: true,
	}
	return result, nil

	return nil, nil
}

func (m *alertService) UpdateOrgCustomizeAlertEnable(ctx context.Context, request *pb.UpdateOrgCustomizeAlertEnableRequest) (*pb.UpdateOrgCustomizeAlertEnableResponse, error) {
	err := m.p.a.UpdateCustomizeAlertEnable(uint64(request.Id), request.Enable)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.UpdateOrgCustomizeAlertEnableResponse{}, nil
}

func (m *alertService) DeleteOrgCustomizeAlert(ctx context.Context, request *pb.DeleteOrgCustomizeAlertRequest) (*pb.DeleteOrgCustomizeAlertResponse, error) {
	data, _ := m.p.a.CustomizeAlert(request.Id)
	err := m.p.a.DeleteCustomizeAlert(request.Id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.DeleteOrgCustomizeAlertResponse{}
	if data != nil {
		resp, err := structpb.NewStruct(map[string]interface{}{
			"name": data.Name,
		})
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		result.Data = structpb.NewStructValue(resp)
		return result, nil
	}
	result.Data = structpb.NewBoolValue(true)
	return result, nil
}

func (m *alertService) QueryDashboardByAlert(ctx context.Context, request *pb.QueryDashboardByAlertRequest) (*pb.QueryDashboardByAlertResponse, error) {
	if request.AlertScope == "" {
		return nil, errors.NewInvalidParameterError("alertScope", "alertScope must not be empty")
	}
	if request.AlertScopeId == "" {
		return nil, errors.NewInvalidParameterError("alertScopeId", "alertScopeId must not be empty")
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alert := &pb.CustomizeAlertDetail{}
	err = json.Unmarshal(data, alert)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	dash, err := adapt.NewDashboard(m.p.a).GenerateDashboardPreView(alert)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.QueryDashboardByAlertResponse{
		Data: dash,
	}
	return result, nil
}

func (m *alertService) QueryAlertRule(ctx context.Context, request *pb.QueryAlertRuleRequest) (*pb.QueryAlertRuleResponse, error) {
	lang := apis.Language(ctx)
	data, err := m.p.a.QueryAlertRule(lang, request.Scope, request.ScopeId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.QueryAlertRuleResponse{
		Data: data,
	}
	return result, nil
}

func (m *alertService) QueryAlert(ctx context.Context, request *pb.QueryAlertRequest) (*pb.QueryAlertsResponse, error) {
	data, err := m.p.a.QueryAlert(apis.Language(ctx), request.Scope, request.ScopeId, uint64(request.PageNo), uint64(request.PageSize))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if data == nil {
		data = make([]*pb.Alert, 0)
	}
	total, err := m.p.a.CountAlert(request.Scope, request.ScopeId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.QueryAlertsResponse{
		Data: &pb.QueryAlertsData{
			List:  data,
			Total: int64(total),
		},
	}
	return result, nil
}

func (m *alertService) GetAlert(ctx context.Context, request *pb.GetAlertRequest) (*pb.GetAlertResponse, error) {
	lang := apis.Language(ctx)
	data, err := m.p.a.GetAlert(lang, uint64(request.Id))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.GetAlertResponse{
		Data: data,
	}
	return result, nil
}

func (m *alertService) GetAlertDetail(ctx context.Context, request *pb.GetAlertDetailRequest) (*pb.GetAlertDetailResponse, error) {
	lang := apis.Language(ctx)
	data, err := m.p.a.GetAlertDetail(lang, uint64(request.Id))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &pb.GetAlertDetailResponse{}
	result.Data = data
	return result, nil
}

func (m *alertService) UpdateAlert(ctx context.Context, request *pb.UpdateAlertRequest) (*pb.UpdateAlertResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertRequest := &pb.Alert{}
	err = json.Unmarshal(data, alertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if err := m.p.checkAlert(alertRequest); err != nil {
		return nil, fmt.Errorf("check alert is failed err is %s", err)
	}
	orgID := apis.GetHeader(ctx, "Org-ID")
	org, err := m.p.bdl.GetOrg(orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertRequest.Attributes["org_name"] = structpb.NewStringValue(org.Name)
	if err := m.p.a.UpdateAlert(request.Id, alertRequest); err != nil {
		if adapt.IsInvalidParameterError(err) {
			return nil, errors.NewInternalServerError(err)
		}
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.UpdateAlertResponse{}, nil
}

func (m alertService) UpdateAlertEnable(ctx context.Context, request *pb.UpdateAlertEnableRequest) (*pb.UpdateAlertEnableResponse, error) {
	err := m.p.a.UpdateAlertEnable(uint64(request.Id), request.Enable)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.UpdateAlertEnableResponse{}, nil
}

func (m *alertService) DeleteAlert(ctx context.Context, request *pb.DeleteAlertRequest) (*pb.DeleteAlertResponse, error) {
	lang := apis.Language(ctx)
	data, _ := m.p.a.GetAlert(lang, uint64(request.Id))
	err := m.p.a.DeleteAlert(uint64(request.Id))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if data != nil {
		return &pb.DeleteAlertResponse{
			Data: map[string]*structpb.Value{
				"name": structpb.NewStringValue(data.Name),
			},
		}, nil
	}
	return &pb.DeleteAlertResponse{}, nil
}

func (m *alertService) QueryOrgAlertRule(ctx context.Context, request *pb.QueryOrgAlertRuleRequest) (*pb.QueryOrgAlertRuleResponse, error) {
	orgID := apis.GetOrgID(ctx)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("orgId", "orgId is invalidate")
	}
	lang := apis.Language(ctx)
	data, err := m.p.a.QueryOrgAlertRule(lang, id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.QueryOrgAlertRuleResponse{Data: data}, nil
}

func (m *alertService) QueryOrgAlert(ctx context.Context, request *pb.QueryOrgAlertRequest) (*pb.QueryOrgAlertResponse, error) {
	orgID := apis.GetOrgID(ctx)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("orgId", "orgId is invalidate")
	}
	lang := apis.Language(ctx)
	data, err := m.p.a.QueryOrgAlert(lang, id, uint64(request.PageNo), uint64(request.PageSize))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	total, err := m.p.a.CountOrgAlert(id)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.QueryOrgAlertResponse{
		Data: &pb.QueryOrgAlertData{
			Total: int64(total),
			List:  data,
		},
	}, nil
}

func (m *alertService) GetOrgAlertDetail(ctx context.Context, request *pb.GetOrgAlertDetailRequest) (*pb.GetOrgAlertDetailResponse, error) {
	orgID := apis.GetOrgID(ctx)
	lang := apis.Language(ctx)
	data, err := m.p.a.GetOrgAlertDetail(lang, uint64(request.Id))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if data == nil {
		return nil, errors.NewNotFoundError("monitor_org_alert")
	}
	if data.AlertScope != "org" || data.AlertScopeId != orgID {
		return nil, nil
	}
	return &pb.GetOrgAlertDetailResponse{
		Data: data,
	}, nil
}

func (m *alertService) checkOrgClusterNames(orgID uint64, clusters []string) bool {
	rels, err := m.p.cmdb.QueryAllOrgClusterRelation()
	if err != nil {
		m.p.L.Errorf("fail to QueryAllOrgClusterRelation(): %s", err)
		return false
	}
	clustersMap := make(map[string]bool)
	for _, item := range rels {
		if item.OrgID == orgID {
			clustersMap[item.ClusterName] = true
		}
	}
	for _, clusterName := range clusters {
		if _, ok := clustersMap[clusterName]; !ok {
			return false
		}
	}
	return true
}

func (m *alertService) UpdateOrgAlert(ctx context.Context, request *pb.UpdateOrgAlertRequest) (*pb.UpdateOrgAlertResponse, error) {
	orgID := apis.GetOrgID(ctx)
	org, err := m.p.bdl.GetOrg(orgID)
	if err != nil {
		return nil, errors.NewInvalidParameterError("orgId", "orgId is invalidate")
	}

	request.Attributes = make(map[string]*structpb.Value)
	orgName := structpb.NewStringValue(org.Name)
	request.Attributes["org_name"] = orgName
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidParameterError("orgId", "orgId is invalidate")
	}
	if len(request.ClusterNames) <= 0 {
		return nil, errors.NewMissingParameterError("cluster names")
	}
	if !m.checkOrgClusterNames(id, request.ClusterNames) {
		return nil, errors.NewPermissionError("monitor_org_alert", "create", "access denied")
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	updateOrgAlertRequest := &pb.Alert{}
	err = json.Unmarshal(data, updateOrgAlertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if err := m.p.a.UpdateOrgAlert(request.Id, updateOrgAlertRequest, orgID); err != nil {
		if adapt.IsInvalidParameterError(err) {
			return nil, errors.NewInvalidParameterError("alertScope", request.AlertScope)
		}
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.UpdateOrgAlertResponse{}, nil
}

func (m *alertService) UpdateOrgAlertEnable(ctx context.Context, request *pb.UpdateOrgAlertEnableRequest) (*pb.UpdateOrgAlertEnableResponse, error) {
	orgID := apis.GetOrgID(ctx)
	if len(orgID) <= 0 {
		return nil, errors.NewInvalidParameterError("Org-ID", "Org-ID not exist")
	}
	err := m.p.a.UpdateOrgAlertEnable(uint64(request.Id), request.Enable, orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.UpdateOrgAlertEnableResponse{}, nil
}

func (m *alertService) DeleteOrgAlert(ctx context.Context, request *pb.DeleteOrgAlertRequest) (*pb.DeleteOrgAlertResponse, error) {
	orgID := apis.GetOrgID(ctx)
	if len(orgID) <= 0 {
		return nil, errors.NewInvalidParameterError("Org-ID", "Org-ID not exist")
	}
	lang := apis.Language(ctx)
	data, _ := m.p.a.GetAlert(lang, uint64(request.Id))
	err := m.p.a.DeleteOrgAlert(uint64(request.Id), orgID)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if data != nil {
		return &pb.DeleteOrgAlertResponse{
			Data: map[string]*structpb.Value{
				"name": structpb.NewStringValue(data.Name),
			},
		}, nil
	}
	return &pb.DeleteOrgAlertResponse{}, nil
}

func (m *alertService) GetAlertRecordAttr(ctx context.Context, request *pb.GetAlertRecordAttrRequest) (*pb.GetAlertRecordAttrResponse, error) {
	lang := apis.Language(ctx)
	data, err := m.p.a.GetAlertRecordAttr(lang, request.Scope)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetAlertRecordAttrResponse{
		Data: data,
	}, nil
}

func (m *alertService) QueryAlertRecord(ctx context.Context, request *pb.QueryAlertRecordRequest) (*pb.QueryAlertRecordResponse, error) {
	if request.PageNo == 0 {
		request.PageNo = 1
	}
	if request.PageSize == 0 {
		request.PageSize = 20
	}
	lang := apis.Language(ctx)
	list, err := m.p.a.QueryAlertRecord(lang, request.Scope, request.ScopeKey,
		request.AlertGroup, request.AlertState, request.AlertType, request.HandleState, request.HandlerId,
		uint(request.PageNo), uint(request.PageSize))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	count, err := m.p.a.CountAlertRecord(request.Scope, request.ScopeKey,
		request.AlertGroup, request.AlertState, request.AlertType, request.HandleState, request.HandlerId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.QueryAlertRecordResponse{
		Data: &pb.ListResult{
			List:  list,
			Total: int64(count),
		},
	}, nil
}

func (m *alertService) GetAlertRecord(ctx context.Context, request *pb.GetAlertRecordRequest) (*pb.GetAlertRecordResponse, error) {
	lang := apis.Language(ctx)
	data, err := m.p.a.GetAlertRecord(lang, request.GroupId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetAlertRecordResponse{
		Data: data,
	}, nil
}

func (m *alertService) QueryAlertHistory(ctx context.Context, request *pb.QueryAlertHistoryRequest) (*pb.QueryAlertHistoryResponse, error) {
	if request.End < request.Start {
		return &pb.QueryAlertHistoryResponse{}, nil
	}
	if request.End == 0 {
		request.End = utils.ConvertTimeToMS(time.Now())
	}
	if request.Start == 0 {
		request.Start = request.End - time.Hour.Milliseconds()
	}
	if request.Limit == 0 {
		request.Limit = 50
	}
	lang := apis.Language(ctx)
	data, err := m.p.a.QueryAlertHistory(lang, request.GroupId, request.Start, request.End, uint(request.Limit))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.QueryAlertHistoryResponse{
		Data: data,
	}, nil
}

func (m *alertService) CreateAlertIssue(ctx context.Context, request *pb.CreateAlertIssueRequest) (*pb.CreateAlertIssueResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	issueCreateRequest := &apistructs.IssueCreateRequest{}
	err = json.Unmarshal(data, issueCreateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if _, err := m.p.a.CreateAlertIssue(request.GroupId, *issueCreateRequest); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.CreateAlertIssueResponse{}, err
}

func (m *alertService) UpdateAlertIssue(ctx context.Context, request *pb.UpdateAlertIssueRequest) (*pb.UpdateAlertIssueResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	issueUpdateRequest := &apistructs.IssueUpdateRequest{}
	err = json.Unmarshal(data, issueUpdateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if err := m.p.a.UpdateAlertIssue(request.GroupId, request.IssueId, *issueUpdateRequest); err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.UpdateAlertIssueResponse{}, nil
}

func (m *alertService) GetOrgAlertRecordAttr(ctx context.Context, request *pb.GetOrgAlertRecordAttrRequest) (*pb.GetOrgAlertRecordAttrResponse, error) {
	lang := apis.Language(ctx)
	data, err := m.p.a.GetOrgAlertRecordAttr(lang)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.GetOrgAlertRecordAttrResponse{
		Data: data,
	}, nil
}

func (m *alertService) QueryOrgAlertRecord(ctx context.Context, request *pb.QueryOrgAlertRecordRequest) (*pb.QueryOrgAlertRecordResponse, error) {
	if request.PageNo == 0 {
		request.PageNo = 1
	}
	if request.PageSize == 0 {
		request.PageSize = 20
	}
	orgID := apis.GetOrgID(ctx)
	lang := apis.Language(ctx)
	list, err := m.p.a.QueryOrgAlertRecord(lang, orgID,
		request.AlertGroup, request.AlertState, request.AlertType, request.HandleState, request.HandlerId,
		uint(request.PageNo), uint(request.PageSize))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	count, err := m.p.a.CountOrgAlertRecord(
		orgID, request.AlertGroup, request.AlertState, request.AlertType, request.HandleState, request.HandlerId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	userIDMap := make(map[string]bool)
	for _, item := range list {
		if item.HandlerId == "" {
			continue
		}
		userIDMap[item.HandlerId] = true
	}
	userIDs := make([]string, 0)
	for key := range userIDMap {
		userIDs = append(userIDs, key)
	}
	return &pb.QueryOrgAlertRecordResponse{
		UserIDs: userIDs,
		Data: &pb.ListResult{
			List:  list,
			Total: int64(count),
		},
	}, nil
}

func (m *alertService) QueryOrgHostsAlertRecord(ctx context.Context, request *pb.QueryOrgHostsAlertRecordRequest) (*pb.QueryOrgAlertRecordResponse, error) {
	request.AlertGroup = make([]string, 0)
	for _, cluster := range request.Clusters {
		for _, hostIP := range cluster.HostIPs {
			request.AlertGroup = append(request.AlertGroup, cluster.ClusterName+"-"+hostIP)
		}
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req := &pb.QueryOrgAlertRecordRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp, err := m.QueryOrgAlertRecord(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.QueryOrgAlertRecordResponse{
		Data:    resp.Data,
		UserIDs: resp.UserIDs,
	}, nil
}

func (m *alertService) GetOrgAlertRecord(ctx context.Context, request *pb.GetOrgAlertRecordRequest) (*pb.GetOrgAlertRecordResponse, error) {
	orgID := apis.GetOrgID(ctx)
	lang := apis.Language(ctx)
	data, err := m.p.a.GetOrgAlertRecord(lang, orgID, request.GroupId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	} else if data == nil {
		return &pb.GetOrgAlertRecordResponse{}, nil
	}
	if data.IssueId != 0 {
		issue, err := m.p.bdl.GetIssue(data.IssueId)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		} else if issue != nil {
			data.ProjectId = issue.ProjectID
		}
	}
	return &pb.GetOrgAlertRecordResponse{
		Data: data,
	}, nil
}

func (m *alertService) QueryOrgAlertHistory(ctx context.Context, request *pb.QueryOrgAlertHistoryRequest) (*pb.QueryOrgAlertHistoryResponse, error) {
	if request.End == 0 {
		request.End = utils.ConvertTimeToMS(time.Now())
	}
	if request.Start == 0 {
		request.Start = request.End - time.Hour.Milliseconds()
	}
	if request.End < request.Start {
		return &pb.QueryOrgAlertHistoryResponse{}, nil
	}
	if request.Limit == 0 {
		request.Limit = 50
	}
	lang := apis.Language(ctx)
	orgID := apis.GetOrgID(ctx)
	data, err := m.p.a.QueryOrgAlertHistory(lang, orgID, request.GroupId, request.Start, request.End, uint(request.Limit))
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.QueryOrgAlertHistoryResponse{
		Data: data,
	}, nil
}

func (m *alertService) CreateOrgAlertIssue(ctx context.Context, request *pb.CreateOrgAlertIssueRequest) (*pb.CreateOrgAlertIssueResponse, error) {
	orgID := apis.GetOrgID(ctx)
	userID := apis.GetUserID(ctx)
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	issueCreateRequest := &apistructs.IssueCreateRequest{}
	err = json.Unmarshal(data, issueCreateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	issueId, err := m.p.a.CreateOrgAlertIssue(orgID, userID, request.GroupId, *issueCreateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.CreateOrgAlertIssueResponse{
		Data: issueId,
	}, nil
}

func (m *alertService) UpdateOrgAlertIssue(ctx context.Context, request *pb.UpdateOrgAlertIssueRequest) (*pb.UpdateOrgAlertIssueResponse, error) {
	orgID := apis.GetOrgID(ctx)
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	issueUpdateRequest := &apistructs.IssueUpdateRequest{}
	err = json.Unmarshal(data, issueUpdateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = m.p.a.UpdateOrgAlertIssue(orgID, request.GroupId, *issueUpdateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &pb.UpdateOrgAlertIssueResponse{}, nil
}
