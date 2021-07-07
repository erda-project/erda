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

	"github.com/gocql/gocql"
	"google.golang.org/protobuf/types/known/structpb"

	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/pkg/common/errors"
)

type traceService struct {
	p *provider
}

func (s *traceService) GetSpans(ctx context.Context, req *pb.GetSpansRequest) (*pb.GetSpansResponse, error) {
	if req.TraceID == "" || req.ScopeID == "" {
		return nil, errors.NewMissingParameterError("traceId or scopeId")
	}
	if req.Limit <= 0 || req.Limit > 1000 {
		req.Limit = 1000
	}
	iter := s.p.cassandraSession.Query("SELECT * FROM spans WHERE trace_id = ? limit ?", req.TraceID, req.Limit).Consistency(gocql.All).Iter()
	var spans []*pb.Span
	for {
		row := make(map[string]interface{})
		if !iter.MapScan(row) {
			break
		}
		var span pb.Span
		span.Id = row["span_id"].(string)
		span.TraceId = row["trace_id"].(string)
		span.OperationName = row["operation_name"].(string)
		span.ParentSpanId = row["parent_span_id"].(string)
		span.StartTime = row["start_time"].(int64)
		span.EndTime = row["end_time"].(int64)
		span.Tags = row["tags"].(map[string]string)
		spans = append(spans, &span)
	}
	return &pb.GetSpansResponse{Data: spans}, nil
}

func (s *traceService) GetTraces(ctx context.Context, req *pb.GetTracesRequest) (*pb.GetTracesResponse, error) {
	if req.ScopeID == "" {
		return nil, errors.NewMissingParameterError("scopeId")
	}
	if req.EndTime <= 0 || req.StartTime <= 0 {
		req.EndTime = time.Now().UnixNano() / 1e6
		h, _ := time.ParseDuration("-1h")
		req.StartTime = time.Now().Add(h).UnixNano() / 1e6
	}
	metricsParams := url.Values{}
	metricsParams.Set("start", strconv.FormatInt(req.StartTime, 10))
	metricsParams.Set("end", strconv.FormatInt(req.EndTime, 10))

	queryParams := make(map[string]*structpb.Value)
	queryParams["terminus_keys"] = structpb.NewStringValue(req.ScopeID)
	queryParams["limit"] = structpb.NewStringValue(strconv.FormatInt(req.Limit, 10))
	var where bytes.Buffer
	if req.ApplicationID > 0 {
		queryParams["applications_ids"] = structpb.NewStringValue(strconv.FormatInt(req.ApplicationID, 10))
		where.WriteString("applications_ids::field=$applications_ids AND ")
	}
	//-1 error, 0 both, 1 success
	statement := fmt.Sprintf("SELECT start_time::field,end_time::field,components::field,"+
		"trace_id::tag,if(gt(errors_sum::field,0),'error','success') FROM trace WHERE %s terminus_keys::field=$terminus_keys "+
		"ORDER BY start_time::field LIMIT %s", where.String(), strconv.FormatInt(req.Limit, 10))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	request := &metricpb.QueryWithInfluxFormatRequest{
		Start:     strconv.FormatInt(req.StartTime, 10),
		End:       strconv.FormatInt(req.EndTime, 10),
		Statement: statement,
		Params:    queryParams,
	}

	response, err := s.p.Metric.QueryWithInfluxFormat(ctx, request)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	rows := response.Results[0].Series[0].Rows
	traces := make([]*pb.Trace, 0, len(rows))
	for _, row := range rows {
		var trace pb.Trace
		values := row.Values
		trace.StartTime = int64(values[0].GetNumberValue() / 1e6)
		trace.Elapsed = math.Abs((values[1].GetNumberValue() - values[0].GetNumberValue()) / 1e6)
		for _, serviceName := range values[2].GetListValue().Values {
			trace.Services = append(trace.Services, serviceName.GetStringValue())
		}
		trace.Id = values[3].GetStringValue()
		status := values[4].GetStringValue()
		if status == "error" && req.Status == 1 {
			continue
		} else if status == "success" && req.Status == -1 {
			continue
		}
		traces = append(traces, &trace)
	}

	return &pb.GetTracesResponse{Data: traces}, nil
}
