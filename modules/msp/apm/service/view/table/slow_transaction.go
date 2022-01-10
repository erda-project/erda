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

package table

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/types/known/structpb"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	slow_transaction "github.com/erda-project/erda/modules/msp/apm/service/common/slow-transaction"
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
	"github.com/erda-project/erda/pkg/common/errors"
)

var (
	slowTransTableColumnOccurTime = &Column{Key: string(slow_transaction.ColumnOccurTime), Name: "Occur Time", Sortable: true}
	slowTransTableColumnDuration  = &Column{Key: string(slow_transaction.ColumnDuration), Name: "Duration", Sortable: true}
	slowTransTableColumnTraceId   = &Column{Key: string(slow_transaction.ColumnTraceId), Name: "Trace Id"}
)

var slowTransactionTableSortFieldSqlMap = map[string]string{
	slowTransTableColumnOccurTime.Key: "timestamp",
	slowTransTableColumnDuration.Key:  "elapsed_mean::field",
}

type SlowTransactionTableRow struct {
	OccurTime float64
	Duration  float64
	TraceId   string
}

func (t *SlowTransactionTableRow) GetCells() []*Cell {
	return []*Cell{
		{Key: slowTransTableColumnOccurTime.Key, Value: t.OccurTime},
		{Key: slowTransTableColumnDuration.Key, Value: t.Duration},
		{Key: slowTransTableColumnTraceId.Key, Value: t.TraceId},
	}
}

type SlowTransactionTableBuilder struct {
	*BaseBuildParams
	MinDuration int64
	MaxDuration int64
}

func (t *SlowTransactionTableBuilder) GetBaseBuildParams() *BaseBuildParams {
	return t.BaseBuildParams
}

func (t *SlowTransactionTableBuilder) GetTable(ctx context.Context) (*Table, error) {
	table := &Table{
		Columns: []*Column{columnPath, columnReqCount, columnErrorCount, columnSlowCount, columnAvgDuration},
	}
	pathField := common.GetLayerPathKeys(t.Layer)[0]
	var layerPathParam *structpb.Value
	if t.FuzzyPath {
		layerPathParam = common.NewStructValue(map[string]interface{}{"regex": ".*" + t.LayerPath + ".*"})
	} else {
		layerPathParam = structpb.NewStringValue(t.LayerPath)
	}
	queryParams := map[string]*structpb.Value{
		"terminus_key": structpb.NewStringValue(t.TenantId),
		"service_id":   structpb.NewStringValue(t.ServiceId),
		"layer_path":   layerPathParam,
	}

	// calculate total count
	statement := fmt.Sprintf("SELECT count(%s) "+
		"FROM %s_slow "+
		"WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) "+
		"%s "+
		"%s "+
		"%s ",
		pathField,
		common.GetDataSourceNames(t.Layer),
		common.BuildDurationFilterSql("elapsed_mean::field", t.MinDuration, t.MaxDuration),
		common.BuildServerSideServiceIdFilterSql("$service_id", t.Layer),
		common.BuildLayerPathFilterSql(t.LayerPath, "$layer_path", t.FuzzyPath, t.Layer),
	)
	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(t.StartTime, 10),
		End:       strconv.FormatInt(t.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
		Options: map[string]string{
			"debug": "true",
		},
	}
	response, err := t.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	table.Total = response.Results[0].Series[0].Rows[0].Values[0].GetNumberValue()

	// query list items
	statement = fmt.Sprintf("SELECT "+
		"timestamp, "+
		"elapsed_mean::field, "+
		"trace_id::tag "+
		"FROM %s_slow "+
		"WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) "+
		"%s "+
		"%s "+
		"%s "+
		"ORDER BY %s "+
		"LIMIT %v OFFSET %v",
		common.GetDataSourceNames(t.Layer),
		common.BuildDurationFilterSql("elapsed_mean::field", t.MinDuration, t.MaxDuration),
		common.BuildServerSideServiceIdFilterSql("$service_id", t.Layer),
		common.BuildLayerPathFilterSql(t.LayerPath, "$layer_path", t.FuzzyPath, t.Layer),
		common.GetSortSql(slowTransactionTableSortFieldSqlMap, "elapsed_mean::field DESC", t.OrderBy...),
		t.PageSize,
		(t.PageNo-1)*t.PageSize,
	)
	request = &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(t.StartTime, 10),
		End:       strconv.FormatInt(t.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
	}
	response, err = t.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	for _, row := range response.Results[0].Series[0].Rows {
		transRow := &SlowTransactionTableRow{
			OccurTime: row.Values[0].GetNumberValue(),
			Duration:  row.Values[1].GetNumberValue(),
			TraceId:   row.Values[2].GetStringValue(),
		}
		table.Rows = append(table.Rows, transRow)
	}

	return table, nil
}
