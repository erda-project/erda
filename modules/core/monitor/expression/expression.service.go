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

	"github.com/erda-project/erda-proto-go/core/monitor/expression/pb"
	"github.com/erda-project/erda/modules/core/monitor/expression/model"
)

var (
	Expressions      []*model.Expression
	MetricExpression []*pb.Expression
	TemplateIndex    map[string][]*model.Template
	ExpressionIndex  map[string]*model.Expression
	ExpressionConfig map[string]*model.ExpressionConfig
	Templates        []*model.Template

	OrgAlertType          []string
	MicroServiceAlertType []string
)

const (
	Org          = "org"
	MicroService = "micro_service"
)

type expressionService struct {
	p *provider
}

func (e *expressionService) GetAlertExpressions(ctx context.Context, request *pb.GetExpressionsRequest) (*pb.GetExpressionsResponse, error) {
	alertExpressions, err := e.p.alertDB.GetAllAlertExpression(request.PageNo, request.PageSize)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(alertExpressions)
	if err != nil {
		return nil, err
	}
	alertExpressionArr := make([]*pb.Expression, 0)
	err = json.Unmarshal(data, &alertExpressionArr)
	if err != nil {
		return nil, err
	}
	return &pb.GetExpressionsResponse{
		Data: alertExpressionArr,
	}, nil
}

func (e *expressionService) GetMetricExpressions(ctx context.Context, request *pb.GetMetricExpressionsRequest) (*pb.GetMetricExpressionsResponse, error) {
	result := &pb.GetMetricExpressionsResponse{}
	if request.PageNo <= 1 {
		if request.PageSize > int64(len(MetricExpression)) {
			request.PageSize = int64(len(MetricExpression))
		}
		result.Data = MetricExpression[:request.PageSize]
	} else {
		if (request.PageNo-1)*request.PageSize >= int64(len(MetricExpression)) {
			return result, nil
		} else if request.PageNo*request.PageSize > int64(len(MetricExpression)) {
			result.Data = MetricExpression[(request.PageNo-1)*request.PageSize:]
		} else {
			result.Data = MetricExpression[(request.PageNo-1)*request.PageSize : request.PageNo*request.PageSize]
		}
	}
	return result, nil
}

func (e *expressionService) GetAlertNotifies(ctx context.Context, request *pb.GetAlertNotifiesRequest) (*pb.GetAlertNotifiesResponse, error) {
	alertNotifies, err := e.p.alertNotifyDB.QueryAlertNotify(request.PageNo, request.PageNo)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(alertNotifies)
	if err != nil {
		return nil, err
	}
	alertNotifyArr := make([]*pb.AlertNotify, 0)
	err = json.Unmarshal(data, &alertNotifyArr)
	if err != nil {
		return nil, err
	}
	return &pb.GetAlertNotifiesResponse{
		Data: alertNotifyArr,
	}, nil
}

func (e *expressionService) GetTemplates(ctx context.Context, request *pb.GetTemplatesRequest) (*pb.GetTemplatesResponse, error) {
	result := &pb.GetTemplatesResponse{
		Data: make([]*pb.AlertTemplate, 0),
	}
	var data []byte
	var err error
	if request.PageNo <= 1 {
		if request.PageSize > int64(len(Templates)) {
			request.PageSize = int64(len(Templates))
		}
		data, err = json.Marshal(Templates[:request.PageSize])
	} else {
		if (request.PageNo-1)*request.PageSize >= int64(len(Templates)) {
			return result, nil
		}
		if request.PageNo*request.PageSize > int64(len(Templates)) {
			data, err = json.Marshal(Templates[(request.PageNo-1)*request.PageSize:])
		} else {
			data, err = json.Marshal(Templates[(request.PageNo-1)*request.PageSize : request.PageNo*request.PageSize])
		}
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &result.Data)
	if err != nil {
		return nil, err
	}
	return result, nil
}
