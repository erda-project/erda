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
	SystemExpressions map[string][]*model.Expression
	SystemTemplate    []*model.Template
)

type expressionService struct {
	p *provider
}

func (e *expressionService) GetAllAlertEnabledExpression(ctx context.Context, request *pb.GetAllAlertEnabledExpressionRequest) (*pb.GetAllAlertEnabledExpressionResponse, error) {
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
	result := &pb.GetAllAlertEnabledExpressionResponse{}
	if request.PageNo <= 1 {
		if request.PageSize > int64(len(alertExpressionArr)) {
			request.PageSize = int64(len(alertExpressionArr))
		}
		result.Data = alertExpressionArr[:request.PageSize]
	} else {
		if (request.PageNo-1)*request.PageSize >= int64(len(alertExpressionArr)) {
			return result, nil
		}
		if request.PageNo*request.PageSize > int64(len(alertExpressionArr)) {
			result.Data = alertExpressionArr[(request.PageNo-1)*request.PageSize:]
		} else {
			result.Data = alertExpressionArr[(request.PageNo-1)*request.PageSize : request.PageNo*request.PageSize]
		}
	}
	return result, nil
}

func (e *expressionService) GetAllMetricEnabledExpression(ctx context.Context, request *pb.GetAllMetricEnabledExpressionRequest) (*pb.GetAllMetricEnabledExpressionResponse, error) {
	metricExpressions, err := e.p.metricDB.GetAllMetricExpression()
	if err != nil {
		return nil, err
	}
	metricExpressionArr := make([]*pb.Expression, 0)
	data, err := json.Marshal(metricExpressions)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &metricExpressionArr)
	if err != nil {
		return nil, err
	}
	result := &pb.GetAllMetricEnabledExpressionResponse{}
	if request.PageNo <= 1 {
		if request.PageSize > int64(len(metricExpressionArr)) {
			request.PageSize = int64(len(metricExpressionArr))
		}
		result.Data = metricExpressionArr[:request.PageSize]
	} else {
		if (request.PageNo-1)*request.PageSize >= int64(len(metricExpressionArr)) {
			return result, nil
		}
		if request.PageNo*request.PageSize > int64(len(metricExpressionArr)) {
			result.Data = metricExpressionArr[(request.PageNo-1)*request.PageSize:]
		} else {
			result.Data = metricExpressionArr[(request.PageNo-1)*request.PageSize : request.PageNo*request.PageSize]
		}
	}
	return result, nil
}

func (e *expressionService) GetAllAlertTemplate(ctx context.Context, request *pb.GetAllAlertTemplateRequest) (*pb.GetAllAlertTemplateResponse, error) {
	result := &pb.GetAllAlertTemplateResponse{
		Data: make([]*pb.AlertTemplate, 0),
	}
	var data []byte
	var err error
	if request.PageNo <= 1 {
		if request.PageSize > int64(len(SystemTemplate)) {
			request.PageSize = int64(len(SystemTemplate))
		}
		data, err = json.Marshal(SystemTemplate[:request.PageSize])
	} else {
		if (request.PageNo-1)*request.PageSize >= int64(len(SystemTemplate)) {
			return result, nil
		}
		if request.PageNo*request.PageSize > int64(len(SystemTemplate)) {
			data, err = json.Marshal(SystemTemplate[(request.PageNo-1)*request.PageSize:])
		} else {
			data, err = json.Marshal(SystemTemplate[(request.PageNo-1)*request.PageSize : request.PageNo*request.PageSize])
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
