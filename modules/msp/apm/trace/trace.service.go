// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package trace

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/erda-project/erda/modules/monitor/trace/query"
	"github.com/erda-project/erda/pkg/common/errors"
)

type traceService struct {
	p       *provider
	metricq metricq.Queryer
	metric  metricpb.MetricServiceServer
	spanq   query.SpanQueryAPI
}

func (s *traceService) GetSpans(ctx context.Context, req *pb.GetSpansRequest) (*pb.GetSpansResponse, error) {
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 1000
	}
	spans := s.spanq.SelectSpans(req.TraceId, req.Limit)
	return &pb.GetSpansResponse{Data: spans}, nil
}
func (s *traceService) GetTraces(ctx context.Context, req *pb.GetTracesRequest) (*pb.GetTracesResponse, error) {
	if req.EndTime <= 0 {
		req.EndTime = time.Now().UnixNano() / 1e6
		h, _ := time.ParseDuration("-1h")
		req.StartTime = time.Now().Add(h).UnixNano() / 1e6
	}
	metricsParams := url.Values{}
	metricsParams.Set("start", strconv.FormatInt(req.StartTime, 10))
	metricsParams.Set("end", strconv.FormatInt(req.EndTime, 10))

	queryParams := make(map[string]interface{})
	queryParams["terminus_keys"] = req.ScopeId
	queryParams["limit"] = strconv.FormatInt(req.Limit, 10)
	var where bytes.Buffer
	if req.ApplicationId > 0 {
		queryParams["applications_ids"] = strconv.FormatInt(req.ApplicationId, 10)
		where.WriteString("applications_ids::field=$applications_ids AND ")
	}
	//-1 error, 0 both, 1 success
	statement := fmt.Sprintf("SELECT start_time::field,end_time::field,components::field,"+
		"trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE %s terminus_keys::field=$terminus_keys "+
		"ORDER BY start_time::field LIMIT %s", where.String(), strconv.FormatInt(req.Limit, 10))
	//s.metric.Query(metricq.InfluxQL, statement, queryParams, metricsParams)
	response, err := s.metricq.Query(metricq.InfluxQL, statement, queryParams, metricsParams)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}

	traces := make([]*pb.Trace, 0, len(response.ResultSet.Rows))
	for _, row := range response.ResultSet.Rows {
		status := row[4].(string)
		if status == "error" && req.Status == 1 {
			continue
		} else if status == "success" && req.Status == -1 {
			continue
		}

		var trace pb.Trace
		trace.Elapsed = math.Abs(row[1].(float64) - row[0].(float64))
		trace.StartTime = int64(row[0].(float64) / 1e6)
		var services []string
		for _, s := range row[2].([]interface{}) {
			services = append(services, s.(string))
		}
		trace.Services = services
		trace.Id = row[3].(string)

		traces = append(traces, &trace)
	}

	return &pb.GetTracesResponse{Data: traces}, nil
}
