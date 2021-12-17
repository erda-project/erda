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

package expression

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/erda-project/erda-infra/providers/i18n"
	alertpb "github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	"github.com/erda-project/erda-proto-go/core/monitor/expression/pb"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/adapt"
	alertdb "github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/core/monitor/expression/model"
	"github.com/erda-project/erda/pkg/common/apis"
)

var SystemExpressions map[string][]*model.Expression

type expressionService struct {
	p *provider
}

func (e *expressionService) GetAllAlertRules(ctx context.Context, request *pb.GetAllAlertRulesRequest) (*pb.GetAllAlertRulesResponse, error) {
	orgID := apis.GetOrgID(ctx)
	id, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return nil, err
	}
	lang := apis.Language(ctx)
	data, err := e.queryAlertRule(lang, id, request.Scope)
	if err != nil {
		return nil, err
	}
	return &pb.GetAllAlertRulesResponse{
		Data: data,
	}, nil
}

func (e *expressionService) queryAlertRule(lang i18n.LanguageCodes, orgID uint64, scope string) (*pb.AllAlertRules, error) {
	rules := SystemExpressions[scope]
	customizeRules, err := e.p.customizeAlertRuleDB.QueryEnabledByScope(scope, string(orgID))
	if err != nil {
		return nil, err
	}
	rulesMap := make(map[string][]*alertpb.AlertRule)
	for _, item := range customizeRules {
		rule, err := adapt.FromCustomizeAlertRule(lang, e.p.t, item)
		if err != nil {
			return nil, err
		}
		rulesMap[item.AlertType] = append(rulesMap[item.AlertType], rule)
	}
	for _, item := range rules {
		alertRule := &alertdb.AlertRule{
			Name:       item.Name,
			AlertScope: item.AlertScope,
			AlertType:  item.AlertType,
			AlertIndex: item.AlertIndex,
			Template:   item.Template,
			Attributes: item.Attributes,
		}
		rule := adapt.FromPBAlertRuleModel(lang, e.p.t, alertRule)
		rulesMap[item.AlertType] = append(rulesMap[item.AlertType], rule)
	}
	var alertTypeRules []*alertpb.AlertTypeRule
	for alertType, rules := range rulesMap {
		alertTypeRules = append(alertTypeRules, &alertpb.AlertTypeRule{
			AlertType: &alertpb.DisplayKey{
				Key:     alertType,
				Display: e.p.t.Text(lang, alertType),
			},
			Rules: rules,
		})
	}
	var operators []*alertpb.Operator
	a := adapt.New(nil, nil, nil, e.p.t, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	for _, op := range a.FunctionOperatorKeys(lang) {
		if op.Type == adapt.OperatorTypeOne {
			operators = append(operators, op)
		}
	}
	return &pb.AllAlertRules{
		AlertRule:  alertTypeRules,
		Windows:    model.WindowKeys,
		Operators:  operators,
		Aggregator: a.AggregatorKeys(lang),
		Silence:    a.NotifySilences(lang),
	}, nil
}

func (e *expressionService) GetAllEnabledExpression(ctx context.Context, request *pb.GetAllEnabledExpressionRequest) (*pb.GetAllEnabledExpressionResponse, error) {
	alertExpressions, err := e.p.alertDB.GetAllAlertExpression()
	if err != nil {
		return nil, err
	}
	alertExpressionArr := make([]*pb.Expression, 0)
	data, err := json.Marshal(alertExpressions)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &alertExpressionArr)
	if err != nil {
		return nil, err
	}
	metricExpressions, err := e.p.metricDB.GetAllMetricExpression()
	if err != nil {
		return nil, err
	}
	metricExpressionArr := make([]*pb.Expression, 0)
	data, err = json.Marshal(metricExpressions)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &metricExpressionArr)
	if err != nil {
		return nil, err
	}
	result := &pb.GetAllEnabledExpressionResponse{
		Data: map[string]*pb.EnabledExpression{
			"alert": {
				Expression: alertExpressionArr,
			},
			"metric": {
				Expression: metricExpressionArr,
			},
		},
	}
	return result, nil
}
