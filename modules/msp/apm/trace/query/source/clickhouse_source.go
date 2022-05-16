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
	"bytes"
	"context"
	"fmt"
	"strings"

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
	tracing struct {
		TraceId   string   `ch:"trace_id"`
		StartTime int64    `ch:"min_start_time"`
		SpanCount uint64   `ch:"span_count"`
		Duration  int64    `ch:"duration"`
		Services  []string `ch:"services"`
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
	specSql := "SELECT distinct(trace_id) AS trace_id,toUnixTimestamp64Nano(min(start_time)) AS min_start_time,count(span_id) AS span_count," +
		"(toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS duration FROM %s %s " +
		"GROUP BY trace_id %s LIMIT %v OFFSET %v"
	var where bytes.Buffer
	// trace id condition
	where.WriteString(fmt.Sprintf("WHERE toUnixTimestamp64Milli(start_time) >= %v AND toUnixTimestamp64Milli(end_time) <= %v ", req.StartTime, req.EndTime))

	if req.TraceID != "" {
		where.WriteString("AND trace_id LIKE %" + req.TraceID + "%) ")
	}
	if req.DurationMin > 0 && req.DurationMax > 0 && req.DurationMin < req.DurationMax {
		where.WriteString(fmt.Sprintf("AND ((toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time)) >= %v "+
			"AND (toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time)) <= %v)", req.DurationMin, req.DurationMax))
	}

	var subSqlBuf bytes.Buffer
	subSqlBuf.WriteString(fmt.Sprintf("SELECT distinct(series_id) FROM %s WHERE (key='terminus_key' AND value = '%s') AND ", SpanMetaTable, req.TenantID))

	if req.ServiceName != "" {
		subSqlBuf.WriteString("(key='service_name' AND value LIKE '%" + req.ServiceName + "%') AND ")
	}

	if req.RpcMethod != "" {
		subSqlBuf.WriteString("(key='rpc_method' AND value LIKE '%" + req.RpcMethod + "%') AND ")
	}

	if req.HttpPath != "" {
		subSqlBuf.WriteString("(key='http_path' AND value LIKE '%" + req.HttpPath + "%') AND ")
	}

	subSql := subSqlBuf.String()
	if strings.HasSuffix(subSql, "AND ") {
		subSql = subSql[:len(subSql)-4]
	}

	where.WriteString(fmt.Sprintf("AND series_id IN (%s)", subSql))

	sql := fmt.Sprintf(specSql, SpanSeriesTable, where.String(), chs.sortConditionStrategy(req.Sort), req.PageSize, (req.PageNo-1)*req.PageSize)

	rows, err := chs.Clickhouse.Client().Query(ctx, sql)
	if err != nil {
		return &pb.GetTracesResponse{}, err
	}
	defer rows.Close()

	traces := make([]*pb.Trace, 0, 10)
	for rows.Next() {
		var t tracing
		tracing := &pb.Trace{}
		if err := rows.ScanStruct(&t); err != nil {
			continue
		}
		tracing.Id = t.TraceId
		tracing.Duration = float64(t.Duration)
		tracing.StartTime = t.StartTime
		tracing.SpanCount = int64(t.SpanCount)
		serviceNames, _ := chs.selectKeyByTraceId(ctx, tracing.Id, "service_name")
		tracing.Services = serviceNames
		traces = append(traces, tracing)
	}
	count := chs.GetTraceCount(ctx, where.String())
	return &pb.GetTracesResponse{PageNo: req.PageNo, PageSize: req.PageSize, Data: traces, Total: count}, nil
}

func (chs *ClickhouseSource) selectKeyByTraceId(ctx context.Context, traceId, key string) ([]string, error) {
	keyMap := make(map[string]struct{}, 10)
	keys := make([]string, 0, 10)
	rows, err := chs.Clickhouse.Client().Query(ctx, fmt.Sprintf("SELECT value FROM %s WHERE series_id IN (SELECT series_id FROM %s WHERE trace_id = &1) AND key = &2", SpanMetaTable, SpanSeriesTable), traceId, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var keyRealValue string
		if err := rows.Scan(&keyRealValue); err != nil {
			continue
		}
		keyMap[keyRealValue] = struct{}{}
	}
	for k := range keyMap {
		keys = append(keys, k)
	}
	return keys, nil
}

func (chs *ClickhouseSource) sortConditionStrategy(sort string) string {
	switch strings.ToLower(sort) {
	case strings.ToLower(pb.SortCondition_TRACE_TIME_DESC.String()):
		return "ORDER BY min_start_time DESC"
	case strings.ToLower(pb.SortCondition_TRACE_TIME_ASC.String()):
		return "ORDER BY min_start_time ASC"
	case strings.ToLower(pb.SortCondition_TRACE_DURATION_DESC.String()):
		return "ORDER BY duration DESC"
	case strings.ToLower(pb.SortCondition_TRACE_DURATION_ASC.String()):
		return "ORDER BY duration ASC"
	case strings.ToLower(pb.SortCondition_SPAN_COUNT_DESC.String()):
		return "ORDER BY span_count DESC"
	case strings.ToLower(pb.SortCondition_SPAN_COUNT_ASC.String()):
		return "ORDER BY span_count ASC"
	default:
		return "ORDER BY min_start_time DESC"
	}
}

func (chs *ClickhouseSource) GetSpans(ctx context.Context, req *pb.GetSpansRequest) []*pb.Span {
	sql := fmt.Sprintf("SELECT org_name,series_id,trace_id,span_id,parent_span_id,toUnixTimestamp64Nano(start_time) AS"+
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

		tags := make(map[string]string, 10)
		sms, err := chs.getSpanMeta(ctx, cs)
		if err != nil {
			return nil
		}
		for _, sm := range sms {
			if "operation_name" == sm.Key {
				cs.OperationName = sm.Key
				continue
			}
			tags[sm.Key] = sm.Value
		}
		chSpanCovertToSpan(span, cs)

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
		var sm spanMeta
		if err := rows.ScanStruct(&sm); err != nil {
			return nil, err
		}
		sms = append(sms, &sm)
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

func (chs *ClickhouseSource) GetTraceCount(ctx context.Context, where string) int64 {

	var count uint64
	sql := fmt.Sprintf("SELECT COUNT(trace_id) FROM %s %s", SpanSeriesTable, where)
	if err := chs.Clickhouse.Client().QueryRow(ctx, sql).Scan(&count); err != nil {
		return 0
	}
	return int64(count)
}
