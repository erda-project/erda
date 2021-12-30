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

package datasources

import (
	"context"

	"github.com/erda-project/erda-proto-go/msp/apm/service/pb"
	"github.com/erda-project/erda/modules/msp/apm/service/view/card"
	"github.com/erda-project/erda/modules/msp/apm/service/view/chart"
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
	"github.com/erda-project/erda/modules/msp/apm/service/view/table"
)

func (p *provider) GetChart(ctx context.Context, chartType pb.ChartType, start, end int64, tenantId, serviceId string, layer common.TransactionLayerType, path string) (interface{}, error) {
	baseChart := &chart.BaseChart{
		StartTime: start,
		EndTime:   end,
		TenantId:  tenantId,
		ServiceId: serviceId,
		Layers:    []common.TransactionLayerType{layer},
		LayerPath: path,
		Metric:    p.Metric,
	}

	data, err := chart.Selector(chartType.String(), baseChart, ctx)
	if err != nil {
		return nil, err
	}

	// todo convert model

	return data, nil
}

func (p *provider) GetCard(ctx context.Context, cardType card.CardType, start, end int64, tenantId, serviceId string, layer common.TransactionLayerType, path string) (interface{}, error) {
	baseCard := &card.BaseCard{
		StartTime: start,
		EndTime:   end,
		TenantId:  tenantId,
		ServiceId: serviceId,
		Layer:     layer,
		LayerPath: path,
		Metric:    p.Metric,
	}

	data, err := card.GetCard(ctx, cardType, baseCard)
	if err != nil {
		return nil, err
	}

	// todo convert model

	return data, nil
}

func (p *provider) GetTable(ctx context.Context, tableType table.TableType, start, end int64, tenantId, serviceId string, layer common.TransactionLayerType, path string, pageNo int, pageSize int) (interface{}, error) {
	baseBuilder := &table.BaseBuilder{
		StartTime: start,
		EndTime:   end,
		TenantId:  tenantId,
		ServiceId: serviceId,
		Layer:     layer,
		LayerPath: path,
		Metric:    p.Metric,
	}

	data, err := table.GetTable(ctx, tableType, baseBuilder)
	if err != nil {
		return nil, err
	}

	// todo convert model

	return data, nil
}
