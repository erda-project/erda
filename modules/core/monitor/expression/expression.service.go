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
	"fmt"

	"github.com/erda-project/erda-proto-go/core/monitor/expression/pb"
	"github.com/erda-project/erda/modules/core/monitor/expression/model"
)

var (
	SystemExpressions map[string][]*model.Expression
	SystemTemplate    []*model.Template
)

type expressionService struct {
	p *provider
}

func (e *expressionService) GetAllAlertTemplate(ctx context.Context, request *pb.GetAllAlertTemplateRequest) (*pb.GetAllAlertTemplateResponse, error) {
	result := &pb.GetAllAlertTemplateResponse{
		Data: make([]*pb.AlertTemplate, 0),
	}
	data, err := json.Marshal(SystemTemplate)
	fmt.Println(string(data))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &result.Data)
	if err != nil {
		return nil, err
	}
	return result, nil
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
