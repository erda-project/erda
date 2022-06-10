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
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace/query/commom/custom"
	"github.com/erda-project/erda/pkg/math"
)

type ClickhouseSource struct {
	Clickhouse clickhouse.Interface
	Log        logs.Logger

	CompatibleSource TraceSource
}

type (
	// spanSeries struct {
	// 	OrgName       string            `ch:"org_name"`
	// 	SeriesId      uint64            `ch:"series_id"`
	// 	TraceId       string            `ch:"trace_id"`
	// 	SpanId        string            `ch:"span_id"`
	// 	ParentSpanId  string            `ch:"parent_span_id"`
	// 	OperationName string            `ch:"operation_name"`
	// 	StartTime     int64             `ch:"start_time"`
	// 	EndTime       int64             `ch:"end_time"`
	// 	Tags          map[string]string `ch:"tags"`
	// }
	// spanMeta struct {
	// 	Key   string `ch:"key"`
	// 	Value string `ch:"value"`
	// }
	tracing struct {
		TraceId   string   `ch:"trace_id"`
		StartTime int64    `ch:"min_start_time"`
		SpanCount uint64   `ch:"span_count"`
		Duration  int64    `ch:"duration"`
		Services  []string `ch:"services"`
	}
	distributionItem struct {
		AvgDuration float64   `ch:"avg_duration"`
		Count       uint64    `ch:"trace_count"`
		Date        time.Time `ch:"date"`
	}
)

const (
	SpanSeriesTable = "monitor.spans_series_all"
	SpanMetaTable   = "monitor.spans_meta_all"
)

const (
	Hour   = time.Hour
	Hour3  = 3 * Hour
	Hour6  = 6 * Hour
	Hour12 = 12 * Hour
	Day    = 24 * Hour
	Day3   = 3 * Day
	Day7   = 7 * Day
	Month  = 30 * Day // just 30 day
	Month3 = 3 * Month
)

func GetInterval(duration int64) (int64, string, int64) {
	count := int64(60) * time.Second.Milliseconds()
	if duration > 0 && duration <= Hour.Milliseconds() {
		interval := Hour.Milliseconds() / count
		return 1, "minute", interval
	} else if duration > Hour.Milliseconds() && duration <= Hour3.Milliseconds() {
		interval := Hour3.Milliseconds() / count
		return 3, "minute", interval
	} else if duration > Hour3.Milliseconds() && duration <= Hour6.Milliseconds() {
		interval := 6 * time.Hour.Milliseconds() / count
		return 6, "minute", interval
	} else if duration > Hour6.Milliseconds() && duration <= Hour12.Milliseconds() {
		interval := 12 * time.Hour.Milliseconds() / count
		return 12, "minute", interval
	} else if duration > Hour12.Milliseconds() && duration <= Day.Milliseconds() {
		interval := Day.Milliseconds() / count
		return 24, "minute", interval
	} else if duration > Day.Milliseconds() && duration <= Day3.Milliseconds() {
		interval := Day3.Milliseconds() / count
		return 72, "minute", interval
	} else if duration > Day3.Milliseconds() && duration <= Day7.Milliseconds() {
		interval := Day7.Milliseconds() / count
		return 168, "minute", interval
	} else if duration > Day7.Milliseconds() && duration <= Month.Milliseconds() {
		interval := Month.Milliseconds() / count
		return 12, "hour", interval
	} else {
		interval := Month3.Milliseconds() / count
		return 36, "hour", interval
	}
}

func (chs *ClickhouseSource) GetTraceReqDistribution(ctx context.Context, model custom.Model) ([]*TraceDistributionItem, error) {
	items := make([]*TraceDistributionItem, 0, 10)
	if chs.CompatibleSource != nil {
		result, err := chs.CompatibleSource.GetTraceReqDistribution(ctx, model)
		if err != nil {
			chs.Log.Error("compatible source query error.")
		}
		items = result
	}

	n, unit, interval := GetInterval(model.EndTime - model.StartTime)
	specSql := "SELECT toStartOfInterval(min_start_time, INTERVAL %v %s) AS date,count(trace_id) AS trace_count, " +
		"avg(duration) AS avg_duration FROM (%s) GROUP BY date ORDER BY date WITH FILL STEP %v ;"

	tracingSql := "SELECT distinct(trace_id) AS trace_id,(toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS duration, min(start_time) AS min_start_time FROM %s %s GROUP BY trace_id"

	var where bytes.Buffer
	// trace id condition
	where.WriteString(fmt.Sprintf("WHERE toUnixTimestamp64Milli(start_time) >= %v AND toUnixTimestamp64Milli(end_time) <= %v ", model.StartTime, model.EndTime))

	if model.TraceId != "" {
		where.WriteString("AND trace_id LIKE concat('%','" + model.TraceId + "','%') ")
	}
	if model.DurationMin > 0 && model.DurationMax > 0 && model.DurationMin < model.DurationMax {
		where.WriteString(fmt.Sprintf("AND ((toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time)) >= %v "+
			"AND (toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time)) <= %v)", model.DurationMin, model.DurationMax))
	}
	where.WriteString(fmt.Sprintf("AND series_id IN (%s)", chs.composeFilter(&pb.GetTracesRequest{TenantID: model.TenantId, ServiceName: model.ServiceName, RpcMethod: model.RpcMethod, HttpPath: model.HttpPath})))

	tracingSql = fmt.Sprintf(tracingSql, SpanSeriesTable, where.String())
	sql := fmt.Sprintf(specSql, n, unit, tracingSql, interval)
	rows, err := chs.Clickhouse.Client().Query(ctx, sql)
	if err != nil {
		chs.Log.Error(err)
		return []*TraceDistributionItem{}, err
	}

	for rows.Next() {
		item := &TraceDistributionItem{}
		var i distributionItem
		if err := rows.ScanStruct(&i); err != nil {
			chs.Log.Error(err)
			continue
		}
		item.Date = i.Date.Format("2006-01-02 15:04:05")
		item.Count = i.Count
		item.AvgDuration = math.DecimalPlacesWithDigitsNumber(i.AvgDuration, 2)
		items = append(items, item)
	}
	return items, nil
}

func (chs *ClickhouseSource) GetTraces(ctx context.Context, req *pb.GetTracesRequest) (*pb.GetTracesResponse, error) {
	traces := make([]*pb.Trace, 0, 10)
	if chs.CompatibleSource != nil {
		compatibleTraces, err := chs.CompatibleSource.GetTraces(ctx, req)
		if err != nil {
			chs.Log.Errorf("compatible query error. err: %v", err)
		} else if compatibleTraces.Total > req.PageSize {
			return compatibleTraces, nil
		}
	}

	specSql := "SELECT distinct(trace_id) AS trace_id,toUnixTimestamp64Nano(min(start_time)) AS min_start_time,count(span_id) AS span_count," +
		"(toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS duration FROM %s %s " +
		"GROUP BY trace_id %s LIMIT %v OFFSET %v"
	var where bytes.Buffer
	// trace id condition
	where.WriteString(fmt.Sprintf("WHERE toUnixTimestamp64Milli(start_time) >= %v AND toUnixTimestamp64Milli(end_time) <= %v ", req.StartTime, req.EndTime))

	if req.TraceID != "" {
		where.WriteString("AND trace_id LIKE concat('%','" + req.TraceID + "','%') ")
	}
	if req.DurationMin > 0 && req.DurationMax > 0 && req.DurationMin < req.DurationMax {
		where.WriteString(fmt.Sprintf("AND ((toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time)) >= %v "+
			"AND (toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time)) <= %v)", req.DurationMin, req.DurationMax))
	}

	where.WriteString(fmt.Sprintf("AND series_id GLOBAL IN (%s)", chs.composeFilter(req)))

	sql := fmt.Sprintf(specSql, SpanSeriesTable, where.String(), chs.sortConditionStrategy(req.Sort), req.PageSize, (req.PageNo-1)*req.PageSize)

	rows, err := chs.Clickhouse.Client().Query(ctx, sql)
	if err != nil {
		return &pb.GetTracesResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var t tracing
		tracing := &pb.Trace{}
		if err := rows.ScanStruct(&t); err != nil {
			chs.Log.Error(err)
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

func (chs *ClickhouseSource) composeFilter(req *pb.GetTracesRequest) string {
	var subSqlBuf bytes.Buffer
	subSqlBuf.WriteString(fmt.Sprintf("SELECT distinct(series_id) FROM %s WHERE (series_id global in (select distinct(series_id) from %s where (key = 'terminus_key' AND value = '%s'))) AND ", SpanMetaTable, SpanMetaTable, req.TenantID))

	if req.ServiceName != "" {
		subSqlBuf.WriteString("(series_id global in (select distinct(series_id) from " + SpanMetaTable + " where (key='service_name' AND value LIKE concat('%','" + req.ServiceName + "','%')))) AND ")
	}

	if req.RpcMethod != "" {
		subSqlBuf.WriteString("(series_id global in (select distinct(series_id) from " + SpanMetaTable + " where (key='rpc_method' AND value LIKE concat('%','" + req.RpcMethod + "','%')))) AND ")
	}

	if req.HttpPath != "" {
		subSqlBuf.WriteString("(series_id global in (select distinct(series_id) from " + SpanMetaTable + " where (key='http_path' AND value LIKE concat('%','" + req.HttpPath + "','%')))) AND ")
	}

	subSql := subSqlBuf.String()
	if strings.HasSuffix(subSql, "AND ") {
		subSql = subSql[:len(subSql)-4]
	}
	return subSql
}

func (chs *ClickhouseSource) selectKeyByTraceId(ctx context.Context, traceId, key string) ([]string, error) {
	keyMap := make(map[string]struct{}, 10)
	keys := make([]string, 0, 10)
	sql := fmt.Sprintf("SELECT value FROM %s WHERE series_id IN (SELECT series_id FROM %s WHERE trace_id = $1) AND key = $2", SpanMetaTable, SpanSeriesTable)
	rows, err := chs.Clickhouse.Client().Query(ctx, sql, traceId, key)
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
	spans := make([]*pb.Span, 0, 10)

	if chs.CompatibleSource != nil {
		getSpans := chs.CompatibleSource.GetSpans(ctx, req)
		if getSpans != nil && len(getSpans) > 0 {
			return getSpans
		}
	}

	sql := fmt.Sprintf(`
SELECT org_name,
       series_id,
       trace_id,
       span_id,
       parent_span_id,
       toUnixTimestamp64Nano(start_time) AS start_time,
       toUnixTimestamp64Nano(end_time)   AS end_time,
       tags
FROM %s
WHERE trace_id = $1
ORDER BY %s LIMIT %v`,
		SpanSeriesTable, "start_time", req.Limit)

	rows, err := chs.Clickhouse.Client().Query(ctx, sql, req.TraceID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var cs trace.Series
		if err := rows.ScanStruct(&cs); err != nil {
			return nil
		}

		sms, err := chs.getSpanMeta(ctx, cs)
		if err != nil {
			return nil
		}

		spans = append(spans, mergeAsSpan(cs, sms))
	}
	return spans
}

func (chs *ClickhouseSource) getSpanMeta(ctx context.Context, cs trace.Series) ([]trace.Meta, error) {
	sms := make([]trace.Meta, 0, 10)
	sql := fmt.Sprintf("SELECT key,value FROM %s WHERE series_id = $1", SpanMetaTable)
	rows, err := chs.Clickhouse.Client().Query(ctx, sql, cs.SeriesID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var sm trace.Meta
		if err := rows.ScanStruct(&sm); err != nil {
			return nil, err
		}
		sms = append(sms, sm)
	}
	return sms, nil
}

func mergeAsSpan(cs trace.Series, sms []trace.Meta) *pb.Span {
	span := &pb.Span{}
	tags := make(map[string]string, 10)
	for _, sm := range sms {
		if "operation_name" == sm.Key {
			span.OperationName = sm.Value
			continue
		}
		tags[sm.Key] = sm.Value
	}
	// merge high cardinality tag
	for k, v := range cs.Tags {
		tags[k] = v
	}
	span.Id = cs.SpanId
	span.TraceId = cs.TraceId
	span.ParentSpanId = cs.ParentSpanId
	span.StartTime = cs.StartTime
	span.EndTime = cs.EndTime
	span.Tags = tags
	return span
}

//
// func chSpanCovertToSpan(span *pb.Span, cs spanSeries) {
// 	span.Id = cs.SpanId
// 	span.TraceId = cs.TraceId
// 	span.ParentSpanId = cs.ParentSpanId
// 	span.OperationName = cs.OperationName
// 	span.StartTime = cs.StartTime
// 	span.EndTime = cs.EndTime
// }

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
