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
	"github.com/erda-project/erda/modules/msp/apm/service/view/common"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/erda-project/erda/pkg/math"
)

type ErrorCountChart struct {
	*BaseChart
}

func (errorCount *ErrorCountChart) GetChart(ctx context.Context) (*pb.ServiceChart, error) {
	statement := fmt.Sprintf("SELECT sum(if(eq(error::tag, 'true'),elapsed_count::field,0)) "+
		"FROM %s "+
		"WHERE (target_terminus_key::tag=$terminus_key OR source_terminus_key::tag=$terminus_key) "+
		"%s "+
		"%s "+
		"GROUP BY time(%s)",
		common.GetDataSourceNames(errorCount.Layers...),
		common.BuildServerSideServiceIdFilterSql("$service_id", errorCount.Layers...),
		common.BuildLayerPathFilterSql(errorCount.LayerPath, "$layer_path", errorCount.FuzzyPath, errorCount.Layers...),
		errorCount.Interval)

	var layerPathParam *structpb.Value
	if errorCount.FuzzyPath {
		layerPathParam = common.NewStructValue(map[string]interface{}{"regex": ".*" + errorCount.LayerPath + ".*"})
	} else {
		layerPathParam = structpb.NewStringValue(errorCount.LayerPath)
	}

	queryParams := map[string]*structpb.Value{
		"terminus_key": structpb.NewStringValue(errorCount.TenantId),
		"service_id":   structpb.NewStringValue(errorCount.ServiceId),
		"layer_path":   layerPathParam,
	}
	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(errorCount.StartTime, 10),
		End:       strconv.FormatInt(errorCount.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
	}
	response, err := errorCount.Metric.QueryWithInfluxFormat(ctx, request)
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
		errorRateChart.Value = math.DecimalPlacesWithDigitsNumber(row.Values[1].GetNumberValue(), 2)
		errorRateChart.Dimension = "Error Count"
		errorRateCharts = append(errorRateCharts, errorRateChart)
	}
	return &pb.ServiceChart{Type: pb.ChartType_ErrorCount.String(), View: errorRateCharts}, err
}
