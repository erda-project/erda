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

package chart

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/service/pb"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/math"
)

type SlowCountChart struct {
	*BaseChart
}

func (slowCount *SlowCountChart) GetChart(ctx context.Context) (*pb.ServiceChart, error) {
	slowThreshold, err := slowCount.getSlowThreshold()
	if err != nil {
		return nil, err
	}
	statement := fmt.Sprintf("SELECT sum(if(gt(elapsed_mean::field, $slow_threshold),elapsed_count::field,0)) "+
		"FROM %s "+
		"WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) "+
		"AND target_service_id::tag=$service_id "+
		"%s "+
		"GROUP BY time(%s)",
		slowCount.getDataSourceNames(),
		slowCount.buildLayerPathFilterSql("$layer_path"),
		slowCount.Interval)
	queryParams := map[string]*structpb.Value{
		"terminus_key":   structpb.NewStringValue(slowCount.TenantId),
		"service_id":     structpb.NewStringValue(slowCount.ServiceId),
		"layer_path":     structpb.NewStringValue(slowCount.LayerPath),
		"slow_threshold": structpb.NewNumberValue(slowThreshold),
	}
	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(slowCount.StartTime, 10),
		End:       strconv.FormatInt(slowCount.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
	}
	response, err := slowCount.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	errorRateCharts := make([]*pb.Chart, 0, 10)

	rows := response.Results[0].Series[0].Rows

	for _, row := range rows {
		errorRateChart := new(pb.Chart)
		date := row.Values[0].GetStringValue()
		parse, err := time.ParseInLocation(Layout, date, time.Local)
		if err != nil {
			return nil, err
		}
		timestamp := parse.UnixNano() / int64(time.Millisecond)

		errorRateChart.Timestamp = timestamp
		errorRateChart.Value = math.DecimalPlacesWithDigitsNumber(row.Values[1].GetNumberValue()*1e2, 2)
		errorRateChart.Dimension = "Error Rate"
		errorRateCharts = append(errorRateCharts, errorRateChart)
	}
	return &pb.ServiceChart{Type: pb.ChartType_ErrorRate.String(), View: errorRateCharts}, err
}

func (slowCount *SlowCountChart) getSlowThreshold() (float64, error) {
	if len(slowCount.Layers) != 1 {
		return 0, fmt.Errorf("do not support multi layers")
	}
	switch slowCount.Layers[0] {
	case TransactionLayerHttp:
		return 300000000, nil
	case TransactionLayerRpc:
		return 300000000, nil
	case TransactionLayerCache:
		return 50000000, nil
	case TransactionLayerDb:
		return 100000000, nil
	case TransactionLayerMq:
		return 300000000, nil
	default:
		return 300000000, nil
	}
}
