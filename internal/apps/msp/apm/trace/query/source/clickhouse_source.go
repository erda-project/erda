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
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace/query/commom/custom"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/pkg/math"
)

type ClickhouseSource struct {
	Clickhouse clickhouse.Interface
	Log        logs.Logger
	DebugSQL   bool
	Loader     loader.Interface

	CompatibleSource TraceSource
}

type (
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
	keysValues struct {
		Keys     []string `ch:"keys"`
		Values   []string `ch:"values"`
		SeriesID uint64   `ch:"series_id"`
	}
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

func fromTimestampMilli(ts int64) exp.SQLFunctionExpression {
	return goqu.Func("fromUnixTimestamp64Milli", goqu.Func("toInt64", ts))
}

type filter struct {
	StartTime, EndTime                                           int64 // ms
	DurationMin, DurationMax                                     int64 // nano
	OrgName, TenantID, TraceID, HttpPath, ServiceName, RpcMethod string
}

func buildFilter(sel *goqu.SelectDataset, f filter) *goqu.SelectDataset {
	// force condition
	sel = sel.Where(goqu.Ex{
		"org_name":   f.OrgName,
		"tenant_id":  f.TenantID,
		"start_time": goqu.Op{"gte": fromTimestampMilli(f.StartTime)},
		"end_time":   goqu.Op{"lte": fromTimestampMilli(f.EndTime)},
	})

	// optional condition
	if f.TraceID != "" {
		sel = sel.Where(goqu.Ex{"trace_id": goqu.Op{"LIKE": "%" + f.TraceID + "%"}})
	}

	if f.DurationMin > 0 && f.DurationMax > 0 && f.DurationMin < f.DurationMax {
		sel = sel.Where(goqu.Ex{
			"((toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time))": goqu.Op{"gte": f.DurationMin},
			"(toUnixTimestamp64Nano(end_time) - toUnixTimestamp64Nano(start_time))":  goqu.Op{"lte": f.DurationMax},
		})
	}

	if f.HttpPath != "" {
		sel = sel.Where(goqu.Ex{
			"tag_values[indexOf(tag_keys, 'http_path')": goqu.Op{"LIKE": "%" + f.HttpPath + "%"},
		})
	}
	if f.ServiceName != "" {
		sel = sel.Where(goqu.Ex{
			"tag_values[indexOf(tag_keys, 'service_name')": goqu.Op{"LIKE": "%" + f.ServiceName + "%"},
		})
	}
	if f.RpcMethod != "" {
		sel = sel.Where(goqu.Ex{
			"tag_values[indexOf(tag_keys, 'rpc_method')": goqu.Op{"LIKE": "%" + f.RpcMethod + "%"},
		})
	}
	return sel
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

	orgName := getOrgName(ctx)
	table, _ := chs.Loader.GetSearchTable(orgName)
	subSQL := goqu.From(table)
	subSQL = subSQL.Select(
		goqu.L("distinct(trace_id)").As("trace_id"),
		goqu.L("(toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time)))").As("duration"),
		goqu.L("min(start_time)").As("min_start_time"),
	)
	subSQL = buildFilter(subSQL, filter{
		OrgName:     orgName,
		TenantID:    model.TenantId,
		StartTime:   model.StartTime,
		EndTime:     model.EndTime,
		DurationMin: model.DurationMin,
		DurationMax: model.DurationMax,
		TraceID:     model.TraceId,
		HttpPath:    model.HttpPath,
		ServiceName: model.ServiceName,
		RpcMethod:   model.RpcMethod,
	})
	subSQL = subSQL.GroupBy("trace_id")

	n, unit, interval := GetInterval(model.EndTime - model.StartTime)
	sql := goqu.From(subSQL).Select(
		goqu.L(fmt.Sprintf("toStartOfInterval(min_start_time, INTERVAL %d %s)", n, unit)).As("date"),
		goqu.L("count(trace_id)").As("trace_count"),
		goqu.L("avg(duration)").As("avg_duration"),
	).GroupBy("date").Order(goqu.C("date").Asc())

	sqlstr, err := chs.toSQL(sql)
	if err != nil {
		return nil, fmt.Errorf("sql convert: %w", err)
	}

	sqlstr += fmt.Sprintf(" WITH FILL STEP %d", interval)
	rows, err := chs.Clickhouse.Client().Query(ctx, sqlstr)
	if err != nil {
		return nil, err
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

	orgName := getOrgName(ctx)
	table, _ := chs.Loader.GetSearchTable(orgName)
	sel := goqu.From(table)
	sel = sel.Select(
		goqu.L("distinct(trace_id)").As("trace_id"),
		goqu.L("toUnixTimestamp64Nano(min(start_time))").As("min_start_time"),
		goqu.L("count(span_id)").As("span_count"),
		goqu.L("(toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time)))").As("duration"),
		goqu.L("groupUniqArrayIf( tag_values[indexOf(tag_keys, 'service_name')], has(tag_keys, 'service_name'))").As("services"),
	)
	f := filter{
		OrgName:     orgName,
		TenantID:    req.TenantID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		DurationMin: req.DurationMin,
		DurationMax: req.DurationMax,
		TraceID:     req.TraceID,
		HttpPath:    req.HttpPath,
		ServiceName: req.ServiceName,
		RpcMethod:   req.RpcMethod,
	}
	sel = buildFilter(sel, f).GroupBy("trace_id").Order(goqu.C("min_start_time").Desc()).
		Limit(uint(req.PageSize)).Offset(uint((req.PageNo - 1) * req.PageSize))

	sqlstr, err := chs.toSQL(sel)
	if err != nil {
		return nil, fmt.Errorf("sql convert: %w", err)
	}

	rows, err := chs.Clickhouse.Client().Query(ctx, sqlstr)
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
		tracing.Services = t.Services
		traces = append(traces, tracing)
	}

	count := chs.GetTraceCount(ctx, f)
	return &pb.GetTracesResponse{PageNo: req.PageNo, PageSize: req.PageSize, Data: traces, Total: count}, nil
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
	if chs.CompatibleSource != nil {
		getSpans := chs.CompatibleSource.GetSpans(ctx, req)
		if getSpans != nil && len(getSpans) > 0 {
			return getSpans
		}
	}

	orgName := getOrgName(ctx)
	table, _ := chs.Loader.GetSearchTable(orgName)
	sql := goqu.From(table).Select(
		"trace_id",
		"span_id",
		"parent_span_id",
		"operation_name",
		goqu.L("toUnixTimestamp64Nano(start_time) AS start_time"),
		goqu.L("toUnixTimestamp64Nano(end_time)   AS end_time"),
		"tag_keys",
		"tag_values",
	).Where(goqu.Ex{
		"org_name":   orgName,
		"tenant_id":  req.ScopeID,
		"trace_id":   req.TraceID,
		"start_time": goqu.Op{"gte": req.StartTime * 1000000},
	}).Order(goqu.C("start_time").Asc()).Limit(uint(req.Limit))

	sqlstr, err := chs.toSQL(sql)
	if err != nil {
		chs.Log.Errorf("ToSQL: %s", err)
		return nil
	}
	rows, err := chs.Clickhouse.Client().Query(ctx, sqlstr)
	if err != nil {
		chs.Log.Errorf("querySQL: %s", err)
		return nil
	}
	defer rows.Close()

	spans := make([]*pb.Span, 0, 10)
	for rows.Next() {
		var span trace.TableSpan
		if err := rows.ScanStruct(&span); err != nil {
			chs.Log.Errorf("scan: %s", err)
			return nil
		}
		tags := make(map[string]string, len(span.TagKeys))
		for i := 0; i < len(span.TagKeys); i++ {
			tags[span.TagKeys[i]] = span.TagValues[i]
		}
		spans = append(spans, &pb.Span{
			Id:            span.SpanId,
			TraceId:       span.TraceId,
			OperationName: span.OperationName,
			StartTime:     span.StartTime,
			EndTime:       span.EndTime,
			ParentSpanId:  span.ParentSpanId,
			Tags:          tags,
		})
	}

	return spans
}

func (chs *ClickhouseSource) GetSpanCount(ctx context.Context, traceID string) int64 {
	var count uint64
	orgName := getOrgName(ctx)
	table, _ := chs.Loader.GetSearchTable(orgName)
	sql := goqu.From(table).Select("COUNT(span_id)").Where(goqu.Ex{"trace_id": traceID})
	sqlstr, err := chs.toSQL(sql)
	if err != nil {
		chs.Log.Errorf("GetSpanCount: %s", err)
		return 0
	}
	if err := chs.Clickhouse.Client().QueryRow(ctx, sqlstr).Scan(&count); err != nil {
		return 0
	}
	return int64(count)
}

func (chs *ClickhouseSource) GetTraceCount(ctx context.Context, f filter) int64 {
	var count uint64
	orgName := getOrgName(ctx)
	table, _ := chs.Loader.GetSearchTable(orgName)
	sql := goqu.From(table).Select(goqu.Func("COUNT", "trace_id"))
	sql = buildFilter(sql, f)
	sqlstr, err := chs.toSQL(sql)
	if err != nil {
		chs.Log.Errorf("GetTraceCount: %s", err)
		return 0
	}
	if err := chs.Clickhouse.Client().QueryRow(ctx, sqlstr).Scan(&count); err != nil {
		chs.Log.Errorf("GetTraceCount query: %s", err)
		return 0
	}
	return int64(count)
}

func (chs *ClickhouseSource) toSQL(sql *goqu.SelectDataset) (string, error) {
	sqlstr, _, err := sql.ToSQL()
	if err != nil {
		return "", fmt.Errorf("tosql err: %w", err)
	}
	if chs.DebugSQL {
		fmt.Printf("===Tracing Clickhouse SQL:\n%s\n===Tracing Clickhouse SQL\n", sqlstr)
	}
	return sqlstr, nil
}
