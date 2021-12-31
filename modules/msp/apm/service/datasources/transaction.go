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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ahmetb/go-linq/v3"

	"github.com/erda-project/erda-infra/providers/component-protocol/components/kv"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/linegraph"
	stdtable "github.com/erda-project/erda-infra/providers/component-protocol/components/table"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-proto-go/msp/apm/service/pb"
	"github.com/erda-project/erda/modules/msp/apm/service/common/transaction"
	"github.com/erda-project/erda/modules/msp/apm/service/view/card"
	"github.com/erda-project/erda/modules/msp/apm/service/view/chart"
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
	"github.com/erda-project/erda/modules/msp/apm/service/view/table"
	"github.com/erda-project/erda/pkg/math"
)

func (p *provider) GetChart(ctx context.Context, chartType pb.ChartType, start, end int64, tenantId, serviceId string, layer common.TransactionLayerType, path string) (*linegraph.Data, error) {
	baseChart := &chart.BaseChart{
		StartTime: start,
		EndTime:   end,
		TenantId:  tenantId,
		ServiceId: serviceId,
		Layers:    []common.TransactionLayerType{layer},
		LayerPath: path,
		FuzzyPath: false,
		Metric:    p.Metric,
	}

	data, err := chart.Selector(strings.ToLower(chartType.String()), baseChart, ctx)
	if err != nil {
		return nil, err
	}
	layout := "2006-01-02 15:04:05"

	// convert model
	var xAxis []interface{}
	linq.From(data.View).
		Select(func(i interface{}) interface{} {
			return time.Unix(0, i.(*pb.Chart).Timestamp*1e6).Format(layout)
		}).
		ToSlice(&xAxis)
	dimension := linq.From(data.View).
		Select(func(i interface{}) interface{} { return i.(*pb.Chart).Dimension }).
		First()
	if dimension == nil {
		dimension = data.Type
	}
	var yAxis []interface{}
	linq.From(data.View).
		Select(func(i interface{}) interface{} {
			value := i.(*pb.Chart).Value
			switch strings.ToLower(chartType.String()) {
			case strings.ToLower(pb.ChartType_AvgDuration.String()):
				return math.DecimalPlacesWithDigitsNumber(value/1e6, 2)
			default:
				return value
			}
		}).
		ToSlice(&yAxis)

	line := linegraph.New(p.I18n.Text(ctx.Value(common.LangKey).(i18n.LanguageCodes), strings.ToLower(data.Type)))
	line.SetXAxis(xAxis...)
	line.SetYAxis(dimension.(string), yAxis...)
	line.SubTitle = chart.GetChartUnitDefault(chartType, ctx.Value(common.LangKey).(i18n.LanguageCodes), p.I18n)

	return line, nil
}

func (p *provider) GetCard(ctx context.Context, cardType card.CardType, start, end int64, tenantId, serviceId string, layer common.TransactionLayerType, path string) (*kv.KV, error) {
	baseCard := &card.BaseCard{
		StartTime: start,
		EndTime:   end,
		TenantId:  tenantId,
		ServiceId: serviceId,
		Layer:     layer,
		LayerPath: path,
		FuzzyPath: false,
		Metric:    p.Metric,
	}

	data, err := card.GetCard(ctx, cardType, baseCard)
	if err != nil {
		return nil, err
	}

	pair := &kv.KV{
		Key:    p.I18n.Text(ctx.Value(common.LangKey).(i18n.LanguageCodes), strings.ToLower(data.Name)),
		Value:  strconv.FormatFloat(data.Value, 'f', -1, 64),
		SubKey: data.Unit,
	}
	return pair, nil
}

func (p *provider) GetTable(ctx context.Context, tableType table.TableType, start, end int64, tenantId, serviceId string, layer common.TransactionLayerType, path string, pageNo int, pageSize int, orderby ...*common.Sort) (*stdtable.Table, error) {
	baseBuilder := &table.BaseBuilder{
		StartTime: start,
		EndTime:   end,
		TenantId:  tenantId,
		ServiceId: serviceId,
		Layer:     layer,
		LayerPath: path,
		FuzzyPath: true,
		OrderBy:   orderby,
		Metric:    p.Metric,
		PageNo:    pageNo,
		PageSize:  pageSize,
	}

	data, err := table.GetTable(ctx, tableType, baseBuilder)
	if err != nil {
		return nil, err
	}

	tt := transaction.InitTable(ctx.Value(common.LangKey).(i18n.LanguageCodes), p.I18n)
	tt.Total = uint64(data.Total)
	tt.PageSize = uint64(pageSize)
	tt.PageNo = uint64(pageNo)

	for _, row := range data.Rows {
		stdrow := stdtable.Row{
			CellsMap: map[stdtable.ColumnKey]stdtable.Cell{},
		}
		for _, cell := range row.GetCells() {
			stdrow.CellsMap[stdtable.ColumnKey(cell.Key)] = stdtable.NewTextCell(fmt.Sprintf("%v", cell.Value)).Build()
		}
		tt.Rows = append(tt.Rows, stdrow)
	}

	return &tt, nil
}
