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

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace/query/commom/custom"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/pkg/ckhelper"
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
		TraceId    string   `ch:"trace_id"`
		TraceCount uint64   `ch:"trace_count"`
		StartTime  int64    `ch:"min_start_time"`
		SpanCount  uint64   `ch:"span_count"`
		Duration   int64    `ch:"duration"`
		Services   []string `ch:"services"`
	}
	distributionItem struct {
		AvgDuration float64   `ch:"avg_duration"`
		Count       uint64    `ch:"trace_count"`
		Date        time.Time `ch:"date"`
	}
	spanItem struct {
		StartTime     int64    `ch:"st_nano"` // timestamp nano
		EndTime       int64    `ch:"et_nano"` // timestamp nano
		OrgName       string   `ch:"org_name"`
		TenantId      string   `ch:"tenant_id"`
		TraceId       string   `ch:"trace_id"`
		SpanId        string   `ch:"span_id"`
		ParentSpanId  string   `ch:"parent_span_id"`
		OperationName string   `ch:"operation_name" `
		TagKeys       []string `ch:"tag_keys"`
		TagValues     []string `ch:"tag_values"`
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
		Status:      model.Status,
		Conditions:  model.Conditions,
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
		goqu.L("count(distinct(trace_id))").As("trace_count"),
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
		Status:      req.Status,
		TraceID:     req.TraceID,
		Conditions:  custom.ConvertConditionByPbCondition(req.Conditions),
	}
	sel = buildFilter(sel, f).GroupBy(goqu.L(`"trace_id" WITH TOTALS`)).Order(chs.sortConditionStrategy(req.Sort)).
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

	totals, err := getTraceTotals(rows)
	if err != nil {
		return &pb.GetTracesResponse{}, err
	}

	return &pb.GetTracesResponse{PageNo: req.PageNo, PageSize: req.PageSize, Data: traces, Total: totals}, nil
}

func getTraceTotals(rows driver.Rows) (int64, error) {
	obj := tracing{}
	err := rows.Totals(&obj.TraceId, &obj.TraceCount, &obj.StartTime, &obj.SpanCount, &obj.Duration, &obj.Services)
	if err != nil {
		return 0, fmt.Errorf("cannot get totals: %w", err)
	}
	return int64(obj.TraceCount), nil
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
		goqu.L("toUnixTimestamp64Nano(start_time) AS st_nano"),
		goqu.L("toUnixTimestamp64Nano(end_time)   AS et_nano"),
		"tag_keys",
		"tag_values",
	).Where(goqu.Ex{
		"org_name":  orgName,
		"tenant_id": req.ScopeID,
		"trace_id":  req.TraceID,
		// there may be some timing skewing between different data source
		"start_time": goqu.Op{"gte": ckhelper.FromTimestampMilli(req.StartTime - (15 * time.Minute).Milliseconds())},
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
		var span spanItem
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
	sql := goqu.From(table).Select(goqu.L("COUNT(span_id)")).Where(goqu.Ex{"trace_id": traceID})
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

func (chs *ClickhouseSource) sortConditionStrategy(sort string) exp.OrderedExpression {
	switch strings.ToLower(sort) {
	case strings.ToLower(pb.SortCondition_TRACE_TIME_DESC.String()):
		return goqu.C("min_start_time").Desc()
	case strings.ToLower(pb.SortCondition_TRACE_TIME_ASC.String()):
		return goqu.C("min_start_time").Asc()
	case strings.ToLower(pb.SortCondition_TRACE_DURATION_DESC.String()):
		return goqu.C("duration").Desc()
	case strings.ToLower(pb.SortCondition_TRACE_DURATION_ASC.String()):
		return goqu.C("duration").Asc()
	case strings.ToLower(pb.SortCondition_SPAN_COUNT_DESC.String()):
		return goqu.C("span_count").Desc()
	case strings.ToLower(pb.SortCondition_SPAN_COUNT_ASC.String()):
		return goqu.C("span_count").Asc()
	default:
		return goqu.C("min_start_time").Desc()
	}
}

func (chs *ClickhouseSource) traceStatusConditionStrategy(traceStatus string) string {
	switch traceStatus {
	case strings.ToLower(pb.TraceStatusCondition_TRACE_SUCCESS.String()):
		return "errors_sum::field=0 AND"
	case strings.ToLower(pb.TraceStatusCondition_TRACE_ALL.String()):
		return "errors_sum::field>=0 AND"
	case strings.ToLower(pb.TraceStatusCondition_TRACE_ERROR.String()):
		return "errors_sum::field>0 AND"
	default:
		return "errors_sum::field>=0 AND"
	}
}

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

type filter struct {
	StartTime, EndTime                 int64 // ms
	DurationMin, DurationMax           int64 // nano
	OrgName, TenantID, TraceID, Status string
	Conditions                         []custom.Condition
}

func buildFilter(sel *goqu.SelectDataset, f filter) *goqu.SelectDataset {
	// force condition
	sel = sel.Where(goqu.Ex{
		"org_name":   f.OrgName,
		"tenant_id":  f.TenantID,
		"start_time": goqu.Op{"gte": ckhelper.FromTimestampMilli(f.StartTime)},
		"end_time":   goqu.Op{"lte": ckhelper.FromTimestampMilli(f.EndTime)},
	})

	// optional condition
	if f.TraceID != "" {
		sel = sel.Where(goqu.C("trace_id").Like("%" + f.TraceID + "%"))
	}

	if len(f.Conditions) > 0 {
		for _, condition := range f.Conditions {
			if condition.TraceId != "" {
				if condition.Operator.IsNotEqualOperator() {
					sel = sel.Where(goqu.C("trace_id").NotLike("%" + condition.TraceId + "%"))
				} else {
					sel = sel.Where(goqu.C("trace_id").Like("%" + condition.TraceId + "%"))
				}
			}

			if condition.HttpPath != "" {
				if condition.Operator.IsNotEqualOperator() {
					sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys, 'http_path')]").NotLike("%" + condition.HttpPath + "%"))
				} else {
					sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys, 'http_path')]").Like("%" + condition.HttpPath + "%"))
				}
			}
			if condition.ServiceName != "" {
				if condition.Operator.IsNotEqualOperator() {
					sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys, 'service_name')]").NotLike("%" + condition.ServiceName + "%"))
				} else {
					sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys, 'service_name')]").Like("%" + condition.ServiceName + "%"))
				}
			}
			if condition.RpcMethod != "" {
				if condition.Operator.IsNotEqualOperator() {
					sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys, 'rpc_method')]").NotLike("%" + condition.RpcMethod + "%"))
				} else {
					sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys, 'rpc_method')]").Like("%" + condition.RpcMethod + "%"))
				}
			}
		}
	}

	if f.DurationMin > 0 && f.DurationMax > 0 && f.DurationMin < f.DurationMax {
		sel = sel.Having(goqu.C("duration").Gte(f.DurationMin), goqu.C("duration").Lte(f.DurationMax))
	}

	switch f.Status {
	case strings.ToLower(pb.TraceStatusCondition_TRACE_SUCCESS.String()):
		sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys,'error')]").Neq("true"))
	case strings.ToLower(pb.TraceStatusCondition_TRACE_ERROR.String()):
		sel = sel.Where(goqu.L("tag_values[indexOf(tag_keys,'error')]").Eq("true"))
	default:
	}

	return sel
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
