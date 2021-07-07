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
	"context"
	"encoding/json"
	"fmt"
	"github.com/erda-project/erda-infra/modcom/api"
	monitor "github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	alert "github.com/erda-project/erda-proto-go/msp/apm/alert/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/structpb"
	"strconv"
	"strings"
)

type alertService struct {
	p *provider
}

func (a *alertService) QueryAlertRule(ctx context.Context, request *alert.QueryAlertRuleRequest) (*alert.QueryAlertRuleResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req := &monitor.QueryAlertRuleRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp, err := a.p.Monitor.QueryAlertRule(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &alert.QueryAlertRuleResponse{
		Data: resp.Data,
	}
	return result, nil
}

func (a *alertService) QueryAlert(ctx context.Context, request *alert.QueryAlertRequest) (*alert.QueryAlertResponse, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req := &monitor.QueryAlertRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp, err := a.p.Monitor.QueryAlert(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &alert.QueryAlertResponse{
		Data: &alert.QueryAlertData{
			List:  resp.Data.List,
			Total: resp.Data.Total,
		},
	}, nil
}

func (a *alertService) GetAlert(ctx context.Context, request *alert.GetAlertRequest) (*alert.GetAlertResponse, error) {
	alertDetailRequest := &monitor.GetAlertDetailRequest{
		Id: request.Id,
	}
	resp, err := a.p.Monitor.GetAlertDetail(ctx, alertDetailRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp.Data.Attributes = nil
	if resp.Data.AlertScope != MicroServiceScope || resp.Data.AlertScopeId != request.TenantGroup {
		return nil, errors.NewInternalServerError(err)
	}
	result := &alert.GetAlertResponse{
		Data: resp.Data,
	}
	return result, nil
}

func (a *alertService) CreateAlert(ctx context.Context, request *alert.CreateAlertRequest) (*alert.CreateAlertResponse, error) {
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	monitorInstance, err := a.p.authDb.Monitor.GetInstanceByTk(tk)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertData := &monitor.Alert{}
	alertData.Name = request.Name
	alertData.Rules = request.Rules
	alertData.Notifies = request.Notifies
	alertData.Domain = request.Domain
	alertData.CreateTime = request.CreateTime
	alertData.UpdateTime = request.UpdateTime
	alertData.AlertScope = MicroServiceScope
	alertData.AlertScopeId = request.TenantGroup
	alertData.Attributes = make(map[string]*structpb.Value)
	alertData.Attributes[DiceOrgId] = structpb.NewStringValue(monitorInstance.OrgId)
	alertData.Attributes[Domain] = structpb.NewStringValue(request.Domain)
	alertData.Attributes[TenantGroup] = structpb.NewStringValue(request.TenantGroup)
	alertData.Attributes[ProjectId] = structpb.NewStringValue(monitorInstance.ProjectId)
	alertData.Attributes[TargetProjectId] = structpb.NewStringValue(monitorInstance.ProjectId)
	alertData.Attributes[WORKSPACE] = structpb.NewStringValue(monitorInstance.Workspace)
	alertData.Attributes[TargetWorkspace] = structpb.NewStringValue(monitorInstance.Workspace)
	alertData.Attributes[TK] = structpb.NewStringValue(monitorInstance.TerminusKey)
	alertData.Attributes[TkAlias] = structpb.NewStringValue(monitorInstance.TerminusKey)
	if request.AppIds != nil && len(request.AppIds) > 0 {
		applicationId, err := (&adapt.Adapt{}).StringSliceToValue(request.AppIds)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		alertData.Attributes[ApplicationId] = applicationId
		alertData.Attributes[TargetApplicationId] = applicationId
	}
	alertData.Attributes[DashboardPath] =
		structpb.NewStringValue(fmt.Sprintf(DashboardPathFormat, monitorInstance.ProjectId, monitorInstance.Workspace, request.TenantGroup, tk))
	alertData.Attributes[RecordPath] =
		structpb.NewStringValue(fmt.Sprintf(RecordPathFormat, monitorInstance.ProjectId, monitorInstance.Workspace, request.TenantGroup, tk))
	ma, err := a.AlertToMonitor(alertData)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if response := a.p.CheckAlert(ma); response != nil {
		return nil, errors.NewInternalServerError(response.(api.Response).Err)
	}
	data, err := json.Marshal(ma)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	createAlertRequest := &monitor.CreateAlertRequest{}
	err = json.Unmarshal(data, createAlertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	resp, err := a.p.Monitor.CreateAlert(ctx, createAlertRequest)
	if err != nil {
		return nil, err
	}
	return &alert.CreateAlertResponse{
		Data: &alert.CreateAlertData{
			Id: resp.Data,
		},
	}, nil
}

func (a *alertService) AlertToMonitor(alertData *monitor.Alert) (*monitor.Alert, error) {
	ma := &monitor.Alert{}
	data, err := json.Marshal(alertData)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, ma)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return ma, nil
}

func (a *alertService) UpdateAlert(ctx context.Context, request *alert.UpdateAlertRequest) (*alert.UpdateAlertResponse, error) {
	getAlertRequest := &monitor.GetAlertRequest{
		Id: int64(request.Id),
	}
	resp, err := a.p.Monitor.GetAlert(ctx, getAlertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if MicroServiceScope != resp.Data.AlertScope || request.TenantGroup != resp.Data.AlertScopeId {
		return nil, errors.NewPermissionError("monitor_project_alert", "update", "scope or scopeId is invalidate")
	}
	alertData := &monitor.Alert{}
	alertData.Id = resp.Data.Id
	alertData.Name = request.Name
	alertData.AlertScope = MicroServiceScope
	alertData.AlertScopeId = request.TenantGroup
	alertData.Rules = request.Rules
	alertData.Notifies = request.Notifies
	alertData.CreateTime = request.CreateTime
	alertData.UpdateTime = request.UpdateTime
	alertData.Attributes = request.Attributes
	if request.AppIds != nil && len(request.AppIds) > 0 {
		alertData.Attributes["application_id"], err = a.p.StringSliceToValue(request.AppIds)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
		alertData.Attributes["target_application_id"], err = a.p.StringSliceToValue(request.AppIds)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
	}
	if request.Domain != "" && len(request.Domain) > 0 {
		alertData.Attributes["alert_domain"] = structpb.NewStringValue(request.Domain)
	}
	data, err := json.Marshal(alertData)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	updateAlertRequest := &monitor.UpdateAlertRequest{}
	err = json.Unmarshal(data, updateAlertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	_, err = a.p.Monitor.UpdateAlert(ctx, updateAlertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (a *alertService) UpdateAlertEnable(ctx context.Context, request *alert.UpdateAlertEnableRequest) (*alert.UpdateAlertEnableResponse, error) {
	getAlertRequest := &monitor.GetAlertRequest{
		Id: request.Id,
	}
	resp, err := a.p.Monitor.GetAlert(ctx, getAlertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if MicroServiceScope != resp.Data.AlertScope || request.TenantGroup != resp.Data.AlertScopeId {
		return nil, errors.NewPermissionError("monitor_project_alert", "update", "scope or scopeId is invalidate")
	}
	updateRequest := &monitor.UpdateAlertEnableRequest{}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, updateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	_, err = a.p.Monitor.UpdateAlertEnable(ctx, updateRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (a *alertService) DeleteAlert(ctx context.Context, request *alert.DeleteAlertRequest) (*alert.DeleteAlertResponse, error) {
	getAlertRequest := &monitor.GetAlertRequest{
		Id: request.Id,
	}
	resp, err := a.p.Monitor.GetAlert(ctx, getAlertRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if MicroServiceScope != resp.Data.AlertScope || request.TenantGroup != resp.Data.AlertScopeId {
		return nil, errors.NewPermissionError("monitor_project_alert", "update", "scope or scopeId is invalidate")
	}
	deleteRequest := &monitor.DeleteAlertRequest{
		Id: request.Id,
	}
	deleteResp, err := a.p.Monitor.DeleteAlert(ctx, deleteRequest)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if resp != nil {
		return &alert.DeleteAlertResponse{
			Data: &alert.DeleteAlertData{
				Name: deleteResp.Data["name"].String(),
			},
		}, nil
	}
	return nil, nil
}

func (a *alertService) QueryCustomizeMetric(ctx context.Context, request *alert.QueryCustomizeMetricRequest) (*alert.QueryCustomizeMetricResponse, error) {
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Errorf("get tk failed"))
	}
	req := &monitor.QueryCustomizeMetricRequest{
		Scope:   MicroServiceScope,
		ScopeId: tk,
	}
	resp, err := a.p.Monitor.QueryCustomizeMetric(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Errorf("get metrics failed"))
	}
	tags := a.p.GetMicroServiceFilterTags()
	for _, metric := range resp.Data.Metrics {
		for i := 0; i < len(metric.Tags); i++ {
			tag := metric.Tags[i]
			if _, ok := tags[tag.Tag.Key]; ok {
				metric.Tags = append(metric.Tags[:i], metric.Tags[i+1:]...)
				i--
			}
		}
	}
	resp.Data.NotifySample = NotifyTemplateSample
	lang := apis.Language(ctx)
	if len(lang) > 0 && strings.HasPrefix(strings.ToLower(lang[0].String()), "en") {
		resp.Data.NotifySample = NotifyTemplateSampleEn
	}
	for i := 0; i < len(resp.Data.FunctionOperators); i++ {
		operator := resp.Data.FunctionOperators[i]
		if operator.Type != OperatorTypeOne {
			resp.Data.FunctionOperators = append(resp.Data.FunctionOperators[:i], resp.Data.FunctionOperators[i+1:]...)
			i--
		}
	}
	for i := 0; i < len(resp.Data.FilterOperators); i++ {
		operator := resp.Data.FilterOperators[i]
		if operator.Type != OperatorTypeOne {
			resp.Data.FilterOperators = append(resp.Data.FilterOperators[:i], resp.Data.FilterOperators[i+1:]...)
			i--
		}
	}
	for _, v := range lang {
		if v.Code == "en" {
			resp.Data.NotifySample = adapt.OrgNotifyTemplateSampleEn
			break
		}
	}
	resp.Data.NotifySample = adapt.OrgNotifyTemplateSample
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &alert.QueryCustomizeMetricResponse{
		Data: resp.Data,
	}
	err = json.Unmarshal(data, result.Data)
	return result, nil
}

func (a *alertService) QueryCustomizeNotifyTarget(ctx context.Context, request *alert.QueryCustomizeNotifyTargetRequest) (*alert.QueryCustomizeNotifyTargetResponse, error) {
	req := &monitor.QueryCustomizeNotifyTargetRequest{}
	resp, err := a.p.Monitor.QueryCustomizeNotifyTarget(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return &alert.QueryCustomizeNotifyTargetResponse{
		Data: resp.Data,
	}, nil
}

func (a *alertService) QueryCustomizeAlerts(ctx context.Context, request *alert.QueryCustomizeAlertsRequest) (*alert.QueryCustomizeAlertsResponse, error) {
	if request.PageSize > 100 {
		request.PageSize = 20
	}
	req := &monitor.QueryCustomizeAlertRequest{
		Scope:    MicroServiceScope,
		ScopeId:  request.TenantGroup,
		PageNo:   request.PageNo,
		PageSize: request.PageSize,
	}
	resp, err := a.p.Monitor.QueryCustomizeAlert(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &alert.QueryCustomizeAlertsResponse{
		Data: resp.Data,
	}
	return result, nil
}

func (a *alertService) GetCustomizeAlert(ctx context.Context, request *alert.GetCustomizeAlertRequest) (*alert.GetCustomizeAlertResponse, error) {
	req := &monitor.GetCustomizeAlertDetailRequest{
		Id: request.Id,
	}
	resp, err := a.p.Monitor.GetCustomizeAlertDetail(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if resp.Data.AlertScopeId != request.TenantGroup {
		return nil, errors.NewPermissionError("monitor_project_alert", "update", "scopeId is invalidate")
	}
	result := &alert.GetCustomizeAlertResponse{
		Data: resp.Data,
	}
	return result, nil
}

func (a *alertService) CreateCustomizeAlert(ctx context.Context, request *alert.CreateCustomizeAlertRequest) (*alert.CreateCustomizeAlertResponse, error) {
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Errorf("get tk failed"))
	}
	customizeMetricsReq := &monitor.QueryCustomizeMetricRequest{
		Scope:   MicroServiceScope,
		ScopeId: tk,
	}
	customizeMetrics, err := a.p.Monitor.QueryCustomizeMetric(ctx, customizeMetricsReq)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Errorf("get metric meta failed"))
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertDetail := &monitor.CustomizeAlertDetail{}
	err = json.Unmarshal(data, alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = checkCustomMetric(customizeMetrics.Data, alertDetail)
	if err != nil {
		return nil, err
	}
	if request.AlertType == "" {
		request.AlertType = "micro_service_customize"
	}
	request.AlertScope = MicroServiceScope
	request.AlertScopeId = request.TenantGroup
	request.Attributes = nil

	for _, rule := range request.Rules {
		rule.Attributes = map[string]*structpb.Value{}
		rule.Attributes[Scope] = structpb.NewStringValue(fmt.Sprintf("{{%s}}", tk))
		scopeFilter := monitor.CustomizeAlertRuleFilter{}
		scopeFilter.Tag = "_metric_scope"
		scopeFilter.Operator = OperateEq
		scopeFilter.Value = structpb.NewStringValue(MicroServiceScope)
		rule.Filters = append(rule.Filters, &scopeFilter)

		scopeIDFilter := monitor.CustomizeAlertRuleFilter{}
		scopeIDFilter.Tag = "_metric_scope"
		scopeIDFilter.Operator = OperateEq
		scopeIDFilter.Value = structpb.NewStringValue(tk)
		rule.Filters = append(rule.Filters, &scopeIDFilter)

		scopeApplicationFilter := monitor.CustomizeAlertRuleFilter{}
		scopeApplicationFilter.Tag = "application_id"
		scopeApplicationFilter.Operator = OperateIn
		scopeApplicationFilter.Value = structpb.NewStringValue("$" + "application_id")
		rule.Filters = append(rule.Filters, &scopeApplicationFilter)
	}
	alertDetailData, err := json.Marshal(alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertDetailReq := &monitor.CreateCustomizeAlertRequest{}
	err = json.Unmarshal(alertDetailData, alertDetailReq)
	resp, err := a.p.Monitor.CreateCustomizeAlert(ctx, alertDetailReq)
	if err != nil {
		if adapt.IsAlreadyExistsError(err) {
			return nil, errors.NewAlreadyExistsError("alert")
		}
		return nil, errors.NewInternalServerError(err)
	}
	return &alert.CreateCustomizeAlertResponse{
		Data: &alert.CreateCustomizeAlertData{
			Id: resp.Data,
		},
	}, nil
}

func checkCustomMetric(customMetrics *monitor.CustomizeMetrics, alert *monitor.CustomizeAlertDetail) error {

	if customMetrics == nil {
		return errors.NewInternalServerError(fmt.Errorf("metric meta not exist"))
	}
	metrics, _, functions, aggregations := selectCustomMetrics(customMetrics)

	aliases := make(map[string]string)
	for _, rule := range alert.Rules {
		// check metric
		if metric, ok := metrics[rule.Metric]; !ok {
			return errors.NewNotFoundError("rule.Metric not exist")
		} else {
			tagsMap := make(map[string]*monitor.TagMeta)
			fieldsMap := make(map[string]*monitor.FieldMeta)
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
					return errors.NewInternalServerError(fmt.Errorf("the %s function alias is empty", function.Field))
				}
				if alias, ok := aliases[function.Alias]; ok {
					return errors.NewInternalServerError(fmt.Errorf("alias :%s duplicate", alias))
				} else {
					aliases[alias] = function.Alias
				}

				if field, ok := fieldsMap[function.Field]; !ok {
					return errors.NewInternalServerError(fmt.Errorf("not support rule function field %s", field))
				}
				if aggregation, ok := aggregations[function.Aggregator]; !ok {
					return errors.NewInternalServerError(fmt.Errorf(fmt.Sprintf("not support rule function aggregator %s", aggregation)))
				}
				if operator, ok := functions[function.Operator]; !ok {
					return errors.NewInternalServerError(fmt.Errorf("not support rule function operator %s", operator))
				}

			}

			for _, filter := range rule.Filters {
				if tag, ok := tagsMap[filter.Tag]; !ok {
					return errors.NewInternalServerError(fmt.Errorf("not support rule filter tag %s", tag))
				}
				if operator, ok := functions[filter.Operator]; !ok {
					return errors.NewInternalServerError(fmt.Errorf("not support rule filter operator %s", operator))
				}
			}

			for _, group := range rule.Group {
				if group, ok := tagsMap[group]; !ok {
					return errors.NewInternalServerError(fmt.Errorf("not support rule filter tag %s", group))
				}
			}
			rule.Outputs = append(rule.Outputs, "alert")
		}
	}
	return nil
}

func selectCustomMetrics(customMetric *monitor.CustomizeMetrics) (map[string]*monitor.MetricMeta,
	map[string]*monitor.Operator, map[string]*monitor.Operator, map[string]*monitor.DisplayKey) {

	metricsMap := make(map[string]*monitor.MetricMeta)
	filtersMap := make(map[string]*monitor.Operator)
	functionsMap := make(map[string]*monitor.Operator)
	aggregationsMap := make(map[string]*monitor.DisplayKey)

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

func (a *alertService) UpdateCustomizeAlert(ctx context.Context, request *alert.UpdateCustomizeAlertRequest) (*alert.UpdateCustomizeAlertResponse, error) {
	req := &monitor.GetCustomizeAlertRequest{
		Id: request.Id,
	}
	customAlertResp, err := a.p.Monitor.GetCustomizeAlert(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if customAlertResp.Data.AlertScope != MicroServiceScope || customAlertResp.Data.AlertScopeId != request.TenantGroup {
		return nil, errors.NewPermissionError("monitor_project_alert", "update", "scope or scopeId is invalidate")
	}
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Errorf("get tk failed"))
	}
	customizeMetricsReq := &monitor.QueryCustomizeMetricRequest{
		Scope:   MicroServiceScope,
		ScopeId: tk,
	}
	customizeMetricsResp, err := a.p.Monitor.QueryCustomizeMetric(ctx, customizeMetricsReq)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Errorf("get metric meta failed"))
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertDetail := &monitor.CustomizeAlertDetail{}
	err = json.Unmarshal(data, alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = checkCustomMetric(customizeMetricsResp.Data, alertDetail)
	if err != nil {
		api.Errors.Internal(err)
	}
	alertDetail.Enable = customAlertResp.Data.Enable
	alertDetail.AlertScope = MicroServiceScope
	alertDetail.Attributes = customAlertResp.Data.Attributes
	for _, rule := range alertDetail.Rules {
		rule.Attributes = map[string]*structpb.Value{}
		rule.Attributes[Scope] = structpb.NewStringValue(fmt.Sprintf("{{%s}}", tk))
		scopeFilter := monitor.CustomizeAlertRuleFilter{}
		scopeFilter.Tag = MetricScope
		scopeFilter.Operator = "eq"
		scopeFilter.Value = structpb.NewStringValue(MicroServiceScope)
		rule.Filters = append(rule.Filters, &scopeFilter)

		scopeIDFilter := monitor.CustomizeAlertRuleFilter{}
		scopeIDFilter.Tag = MetricScopeId
		scopeIDFilter.Operator = "eq"
		scopeIDFilter.Value = structpb.NewStringValue(tk)
		rule.Filters = append(rule.Filters, &scopeIDFilter)

		scopeApplicationFilter := monitor.CustomizeAlertRuleFilter{}
		scopeApplicationFilter.Tag = "application_id"
		scopeApplicationFilter.Operator = OperateIn
		scopeApplicationFilter.Value = structpb.NewStringValue("$" + "application_id")
		rule.Filters = append(rule.Filters, &scopeApplicationFilter)
	}
	err = a.p.checkCustomizeAlert(alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err = json.Marshal(alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	updateCustomizeAlertReq := &monitor.UpdateCustomizeAlertRequest{}
	err = json.Unmarshal(data, updateCustomizeAlertReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	_, err = a.p.Monitor.UpdateCustomizeAlert(ctx, updateCustomizeAlertReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (a *alertService) UpdateCustomizeAlertEnable(ctx context.Context, request *alert.UpdateCustomizeAlertEnableRequest) (*alert.UpdateCustomizeAlertEnableResponse, error) {
	req := &monitor.GetCustomizeAlertRequest{
		Id: request.Id,
	}
	resp, err := a.p.Monitor.GetCustomizeAlert(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if resp == nil {
		return nil, errors.NewNotFoundError("id = " + strconv.FormatUint(request.Id, 10))
	}
	if resp.Data.AlertScopeId != request.TenantGroup {
		return nil, errors.NewPermissionError("monitor_project_alert", "update", "scopeId is invalidate")
	}
	updateReq := &monitor.UpdateCustomizeAlertEnableRequest{
		Id:     request.Id,
		Enable: request.Enable,
	}
	_, err = a.p.Monitor.UpdateCustomizeAlertEnable(ctx, updateReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (a *alertService) DeleteCustomizeAlert(ctx context.Context, request *alert.DeleteCustomizeAlertRequest) (*alert.DeleteCustomizeAlertResponse, error) {
	req := &monitor.DeleteCustomizeAlertRequest{
		Id: request.Id,
	}
	resp, err := a.p.Monitor.DeleteCustomizeAlert(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &alert.DeleteCustomizeAlertResponse{
		Data: &alert.DeleteCustomizeAlertData{
			Name: resp.Data,
		},
	}
	return result, nil
}

func (a *alertService) GetAlertRecordAttrs(ctx context.Context, request *alert.GetAlertRecordAttrsRequest) (*alert.GetAlertRecordAttrsResponse, error) {
	req := &monitor.GetAlertRecordAttrRequest{
		Scope: MicroServiceScope,
	}
	resp, err := a.p.Monitor.GetAlertRecordAttr(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if resp == nil {
		return nil, nil
	}
	if resp.Data.AlertType == nil {
		return nil, nil
	}
	resp.Data.AlertType = append(resp.Data.AlertType, &monitor.DisplayKey{
		Key:     "micro_service_customize",
		Display: "微服务自定义",
	})
	result := &alert.GetAlertRecordAttrsResponse{
		Data: resp.Data,
	}
	return result, nil
}

func (a *alertService) GetAlertRecords(ctx context.Context, request *alert.GetAlertRecordsRequest) (*alert.GetAlertRecordsResponse, error) {
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if tk == "" {
		return nil, errors.NewInternalServerError(fmt.Errorf("tenantGroup has no tk"))
	}
	monitorInstance, err := a.p.authDb.Monitor.GetInstanceByTk(tk)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if request.PageNo == 0 || request.PageNo < 1 {
		request.PageNo = 1
	}
	if request.PageSize <= 0 || request.PageSize > 100 {
		request.PageSize = 20
	}
	req := &monitor.QueryAlertRecordRequest{
		Scope:       MicroServiceScope,
		ScopeKey:    request.TenantGroup,
		AlertGroup:  request.AlertGroup,
		AlertState:  request.AlertState,
		AlertType:   request.AlertType,
		HandleState: request.HandleState,
		HandlerId:   request.HandlerId,
		PageNo:      uint64(request.PageNo),
		PageSize:    uint64(request.PageSize),
	}
	resp, err := a.p.Monitor.QueryAlertRecord(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	var projectId int
	if monitorInstance.ProjectId != "" {
		projectId, err = strconv.Atoi(monitorInstance.ProjectId)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
	}
	userIds := make([]string, 0)
	if resp != nil {
		for index, value := range resp.Data.List {
			if value.HandlerId != "" {
				userIds = append(userIds, value.HandlerId)
			}
			if projectId != 0 {
				resp.Data.List[index].ProjectId = uint64(projectId)
			}
		}
	}
	return &alert.GetAlertRecordsResponse{
		Data: &alert.GetAlertRecordsData{
			List:    resp.Data.List,
			Total:   resp.Data.Total,
			UserIds: userIds,
		},
	}, nil
}

func (a *alertService) GetAlertRecord(ctx context.Context, request *alert.GetAlertRecordRequest) (*alert.GetAlertRecordResponse, error) {
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if tk == "" {
		return nil, errors.NewInternalServerError(fmt.Errorf("tenantGroup has no tk"))
	}
	monitorInstance, err := a.p.authDb.Monitor.GetInstanceByTk(tk)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	var projectId int
	if monitorInstance.ProjectId != "" {
		projectId, err = strconv.Atoi(monitorInstance.ProjectId)
		if err != nil {
			return nil, errors.NewInternalServerError(err)
		}
	}
	req := &monitor.GetAlertRecordRequest{
		GroupId: request.GroupId,
	}
	resp, err := a.p.Monitor.GetAlertRecord(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if resp == nil {
		return nil, nil
	}
	if resp.Data.Scope != MicroServiceScope || resp.Data.ScopeKey != request.TenantGroup {
		return nil, errors.NewPermissionError("monitor_project_alert", "delete", "scopeId or scope is invalidate")
	}
	if projectId != 0 {
		resp.Data.ProjectId = uint64(projectId)
	}
	result := &alert.GetAlertRecordResponse{
		Data: resp.Data,
	}
	return result, nil
}

func (a *alertService) GetAlertHistories(ctx context.Context, request *alert.GetAlertHistoriesRequest) (*alert.GetAlertHistoriesResponse, error) {
	req := &monitor.QueryAlertHistoryRequest{
		GroupId: request.GroupId,
		Start:   request.Start,
		End:     request.End,
		Limit:   uint64(request.Limit),
	}
	resp, err := a.p.Monitor.QueryAlertHistory(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &alert.GetAlertHistoriesResponse{
		Data: resp.Data,
	}
	return result, nil
}

func (a *alertService) CreateAlertRecordIssue(ctx context.Context, request *alert.CreateAlertRecordIssueRequest) (*alert.CreateAlertRecordIssueResponse, error) {
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if tk == "" {
		return nil, errors.NewInternalServerError(fmt.Errorf("tenantGroup has no tk"))
	}
	monitorInstance, err := a.p.authDb.Monitor.GetInstanceByTk(tk)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	httpRequest := utils.GetHttpRequest(ctx)
	userId := httpRequest.Header.Get("User-ID")
	getRecordReq := &monitor.GetAlertRecordRequest{
		GroupId: request.GroupId,
	}
	record, err := a.p.Monitor.GetAlertRecord(ctx, getRecordReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if record == nil || record.Data.IssueId != 0 {
		return nil, nil
	}
	if record.Data.Scope != MicroServiceScope || record.Data.ScopeKey != request.TenantGroup {
		return nil, errors.NewPermissionError("monitor_project_alert", "create", "scopeId or scope is invalidate")
	}
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if monitorInstance.ProjectId == "" {
		return nil, errors.NewInternalServerError(fmt.Errorf("monitor has no project id"))
	}
	projectId, err := strconv.Atoi(monitorInstance.ProjectId)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if request.Body == nil {
		request.Body = make(map[string]*structpb.Value)
		request.Body["creator"] = structpb.NewStringValue(userId)
		request.Body["projectID"] = structpb.NewNumberValue(float64(projectId))
	}
	createIssue := &apistructs.IssueCreateRequest{}
	err = mapstructure.Decode(request.Body, createIssue)
	req := &monitor.CreateAlertIssueRequest{}
	data, err := json.Marshal(createIssue)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	req.GroupId = request.GroupId
	_, err = a.p.Monitor.CreateAlertIssue(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Errorf("alert record issue create fail"))
	}
	return nil, nil
}

func (a *alertService) UpdateAlertRecordIssue(ctx context.Context, request *alert.UpdateAlertRecordIssueRequest) (*alert.UpdateAlertRecordIssueResponse, error) {
	req := &monitor.GetAlertRecordRequest{
		GroupId: request.GroupId,
	}
	record, err := a.p.Monitor.GetAlertRecord(ctx, req)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if record == nil || record.Data.IssueId != 0 {
		return nil, nil
	}
	if record.Data.Scope != MicroServiceScope || record.Data.ScopeKey != request.TenantGroup {
		return nil, errors.NewPermissionError("monitor_project_alert", "update", "scopeId or scope is invalidate")
	}
	issueUpdate := &apistructs.IssueUpdateRequest{}
	err = mapstructure.Decode(request.Body, issueUpdate)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	data, err := json.Marshal(issueUpdate)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	updateDataIssueReq := &monitor.UpdateAlertIssueRequest{}
	err = json.Unmarshal(data, updateDataIssueReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	updateDataIssueReq.GroupId = request.GroupId
	_, err = a.p.Monitor.UpdateAlertIssue(ctx, updateDataIssueReq)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	return nil, nil
}

func (a *alertService) DashboardPreview(ctx context.Context, request *alert.DashboardPreviewRequest) (*alert.DashboardPreviewResponse, error) {
	request.AlertScope = MicroServiceScope
	tk, err := a.p.authDb.InstanceTenant.QueryTkByTenantGroup(request.TenantGroup)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	if tk == "" {
		return nil, errors.NewInternalServerError(fmt.Errorf("no monitor"))
	}
	request.AlertScopeId = tk
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	alertDetail := &monitor.QueryDashboardByAlertRequest{}
	err = json.Unmarshal(data, alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	view, err := a.p.Monitor.QueryDashboardByAlert(ctx, alertDetail)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	result := &alert.DashboardPreviewResponse{
		Data: view.Data,
	}
	return result, nil
}
