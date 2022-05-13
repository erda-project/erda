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

package source

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/msp/apm/trace/query/commom/custom"
)

type ClickhouseSource struct {
	Clickhouse clickhouse.Interface
	Log        logs.Logger
}

type (
	spanSeries struct {
		OrgName       string `ch:"org_name"`
		SeriesId      uint64 `ch:"series_id"`
		TraceId       string `ch:"trace_id"`
		SpanId        string `ch:"span_id"`
		ParentSpanId  string `ch:"parent_span_id"`
		OperationName string `ch:"operation_name"`
		StartTime     int64  `ch:"start_time"`
		EndTime       int64  `ch:"end_time"`
	}
	spanMeta struct {
		Key   string `ch:"key"`
		Value string `ch:"value"`
	}
)

const (
	SpanSeriesTable = "monitor.spans_series"
	SpanMetaTable   = "monitor.spans_meta"
)

func (chs *ClickhouseSource) GetTraceReqDistribution(ctx context.Context, model custom.Model) ([]*TraceDistributionItem, error) {
	return nil, nil
}

func (chs *ClickhouseSource) GetTraces(ctx context.Context, req *pb.GetTracesRequest) (*pb.GetTracesResponse, error) {
	return nil, nil
}

func (chs *ClickhouseSource) GetSpans(ctx context.Context, req *pb.GetSpansRequest) []*pb.Span {
	sql := fmt.Sprintf("SELECT org_name,series_id,trace_id,span_id,parent_span_id,operation_name,toUnixTimestamp64Nano(start_time) AS"+
		" start_time,toUnixTimestamp64Nano(end_time) AS end_time FROM %s WHERE trace_id = $1 ORDER BY %s LIMIT %v",
		SpanSeriesTable, "start_time", req.Limit)

	rows, err := chs.Clickhouse.Client().Query(ctx, sql, req.TraceID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	spans := make([]*pb.Span, 0, 10)

	for rows.Next() {
		span := &pb.Span{}
		var cs spanSeries
		if err := rows.ScanStruct(&cs); err != nil {
			return nil
		}
		chSpanCovertToSpan(span, cs)

		tags := make(map[string]string, 10)
		sms, err := chs.getSpanMeta(ctx, cs)
		if err != nil {
			return nil
		}
		for _, sm := range sms {
			tags[sm.Key] = sm.Value
		}

		span.Tags = tags
		spans = append(spans, span)
	}
	return spans
}

func (chs *ClickhouseSource) getSpanMeta(ctx context.Context, cs spanSeries) ([]*spanMeta, error) {
	sms := make([]*spanMeta, 0, 10)
	sql := fmt.Sprintf("SELECT key,value FROM %s WHERE series_id = $1", SpanMetaTable)
	rows, err := chs.Clickhouse.Client().Query(ctx, sql, cs.SeriesId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var sm *spanMeta
		if err := rows.ScanStruct(sm); err != nil {
			return nil, err
		}
		sms = append(sms, sm)
	}
	return sms, nil
}

func chSpanCovertToSpan(span *pb.Span, cs spanSeries) {
	span.Id = cs.SpanId
	span.TraceId = cs.TraceId
	span.ParentSpanId = cs.ParentSpanId
	span.OperationName = cs.OperationName
	span.StartTime = cs.StartTime
	span.EndTime = cs.EndTime
}

func (chs *ClickhouseSource) GetSpanCount(ctx context.Context, traceID string) int64 {
	var count uint64
	sql := fmt.Sprintf("SELECT COUNT(span_id) FROM %s WHERE trace_id = $1", SpanSeriesTable)
	if err := chs.Clickhouse.Client().QueryRow(ctx, sql, traceID).Scan(&count); err != nil {
		return 0
	}
	return int64(count)
}
