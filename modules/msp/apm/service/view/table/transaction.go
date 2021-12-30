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
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/time"
)

var (
	columnPath        = &Column{Key: "path", Name: "Name"}
	columnReqCount    = &Column{Key: "reqCount", Name: "Req Count"}
	columnErrorCount  = &Column{Key: "errorCount", Name: "Error Count"}
	columnSlowCount   = &Column{Key: "slowCount", Name: "Slow Count"}
	columnAvgDuration = &Column{Key: "avgDuration", Name: "Avg Duration"}
)

var TransactionTableSortFieldSqlMap = map[string]string{
	columnReqCount.Key:    "sum(elapsed_count::field)",
	columnErrorCount.Key:  "count(error::tag)",
	columnSlowCount.Key:   "sum(if(gt(elapsed_mean::field, $slow_threshold),elapsed_count::field,0))",
	columnAvgDuration.Key: "format_duration(avg(elapsed_mean::field),'',2)",
}

type TransactionTableRow struct {
	Path        string
	ReqCount    float64
	ErrorCount  float64
	SlowCount   float64
	AvgDuration string
}

func (t *TransactionTableRow) GetCells() []*Cell {
	return []*Cell{
		{Key: columnPath.Key, Value: t.Path},
		{Key: columnReqCount.Key, Value: t.ReqCount},
		{Key: columnErrorCount.Key, Value: t.ErrorCount},
		{Key: columnSlowCount.Key, Value: t.SlowCount},
		{Key: columnAvgDuration.Key, Value: t.AvgDuration},
	}
}

type TransactionTableBuilder struct {
	*BaseBuilder
}

func (t *TransactionTableBuilder) GetTable(ctx context.Context) (*Table, error) {
	table := &Table{
		Columns: []*Column{columnPath, columnReqCount, columnErrorCount, columnSlowCount, columnAvgDuration},
	}
	pathField := common.GetLayerPathKeys(t.Layer)[0]
	queryParams := map[string]*structpb.Value{
		"terminus_key":   structpb.NewStringValue(t.TenantId),
		"service_id":     structpb.NewStringValue(t.ServiceId),
		"layer_path":     common.NewStructValue(map[string]interface{}{"regex": ".*" + t.LayerPath + ".*"}),
		"slow_threshold": structpb.NewNumberValue(common.GetSlowThreshold(t.Layer)),
	}

	// calculate total count
	statement := fmt.Sprintf("SELECT DISTINCT(%s) "+
		"FROM %s "+
		"WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) "+
		"AND target_service_id::tag=$service_id "+
		"%s "+
		pathField,
		common.GetDataSourceNames(t.Layer),
		common.BuildLayerPathFilterSql(t.LayerPath, "$layer_path", t.Layer))
	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(t.StartTime, 10),
		End:       strconv.FormatInt(t.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
	}
	response, err := t.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	table.Total = response.Results[0].Series[0].Rows[0].Values[0].GetNumberValue()

	// query list items
	statement = fmt.Sprintf("SELECT "+
		"%s as path,"+
		"sum(elapsed_count::field) AS reqCount,"+
		"count(error::tag) AS errorCount,"+
		"sum(if(gt(elapsed_mean::field, $slow_threshold),elapsed_count::field,0)) AS slowCount,"+
		"format_duration(avg(elapsed_mean::field),'',2) AS avgDuration "+
		"FROM %s "+
		"WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) "+
		"AND target_service_id::tag=$service_id "+
		"%s "+
		"GROUP BY %s "+
		"ORDER BY %s "+
		"LIMIT %v OFFSET %v",
		pathField,
		common.GetDataSourceNames(t.Layer),
		common.BuildLayerPathFilterSql(t.LayerPath, "$layer_path", t.Layer),
		pathField,
		common.GetSortSql(TransactionTableSortFieldSqlMap, "sum(elapsed_count::field) DESC"),
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
		duration, unit := time.AutomaticConversionUnit(row.Values[4].GetNumberValue())
		transRow := &TransactionTableRow{
			Path:        row.Values[0].GetStringValue(),
			ReqCount:    row.Values[1].GetNumberValue(),
			ErrorCount:  row.Values[2].GetNumberValue(),
			SlowCount:   row.Values[3].GetNumberValue(),
			AvgDuration: fmt.Sprintf("%v%s", duration, unit),
		}
		table.Rows = append(table.Rows, transRow)
	}

	return table, nil
}
