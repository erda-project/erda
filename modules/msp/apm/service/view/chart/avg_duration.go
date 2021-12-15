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
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/service/pb"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/math"
	"google.golang.org/protobuf/types/known/structpb"
	"strconv"
	"time"
)

type AvgDurationChart struct {
	*BaseChart
}

func (avgDuration *AvgDurationChart) GetChart(ctx context.Context) (*pb.ServiceChart, error) {
	statement := fmt.Sprintf("SELECT sum(elapsed_sum::field)/sum(count_sum::field) "+
		"FROM application_http_service "+
		"WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) "+
		"AND target_service_id::tag=$service_id "+
		"GROUP BY time(%s)", avgDuration.Interval)
	queryParams := map[string]*structpb.Value{
		"terminus_key": structpb.NewStringValue(avgDuration.TenantId),
		"service_id":   structpb.NewStringValue(avgDuration.ServiceId),
	}
	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(avgDuration.StartTime, 10),
		End:       strconv.FormatInt(avgDuration.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
	}
	response, err := avgDuration.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	avgDurationCharts := make([]*pb.Chart, 0, 10)

	rows := response.Results[0].Series[0].Rows

	for _, row := range rows {
		avgDurationChart := new(pb.Chart)
		date := row.Values[0].GetStringValue()
		parse, err := time.ParseInLocation(Layout, date, time.Local)
		if err != nil {
			return nil, err
		}
		timestamp := parse.UnixNano() / int64(time.Millisecond)

		avgDurationChart.Timestamp = timestamp

		avgDurationChart.Value = math.DecimalPlacesWithDigitsNumber(row.Values[1].GetNumberValue(), 2)

		avgDurationCharts = append(avgDurationCharts, avgDurationChart)
	}
	return &pb.ServiceChart{Type: pb.ChartType_AvgDuration.String(), View: avgDurationCharts}, err
}
