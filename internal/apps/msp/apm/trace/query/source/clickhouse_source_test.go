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
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"

	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda/internal/apps/msp/apm/trace/query/commom/custom"

	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
)

func Test_GetInterval(t *testing.T) {
	type args struct {
		duration int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"case1", args{duration: time.Hour.Milliseconds()}, 1},
		{"case2", args{duration: time.Hour.Milliseconds() * 3}, 3},
		{"case3", args{duration: time.Hour.Milliseconds() * 6}, 6},
		{"case4", args{duration: time.Hour.Milliseconds() * 12}, 12},
		{"case5", args{duration: time.Hour.Milliseconds() * 24}, 24},
		{"case6", args{duration: time.Hour.Milliseconds() * 72}, 72},
		{"case7", args{duration: time.Hour.Milliseconds() * 168}, 168},
		{"case8", args{duration: time.Hour.Milliseconds() * 168 * 3}, 12},
		{"case9", args{duration: time.Hour.Milliseconds() * 168 * 5}, 36},
		{"case10", args{duration: time.Hour.Milliseconds() * 168 * 10}, 36},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, _, _ := GetInterval(tt.args.duration)
			if got != tt.want {
				t.Errorf("GetInterval() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClickhouseSource_sortConditionStrategy(t *testing.T) {
	type args struct {
		sort string
	}
	tests := []struct {
		name string
		args args
		want exp.OrderedExpression
	}{
		{"case1", args{sort: ""}, goqu.C("min_start_time").Desc()},
		{"case2", args{sort: "TRACE_TIME_DESC"}, goqu.C("min_start_time").Desc()},
		{"case3", args{sort: "TRACE_TIME_ASC"}, goqu.C("min_start_time").Asc()},
		{"case4", args{sort: "TRACE_DURATION_DESC"}, goqu.C("duration").Desc()},
		{"case5", args{sort: "TRACE_DURATION_ASC"}, goqu.C("duration").Asc()},
		{"case6", args{sort: "SPAN_COUNT_DESC"}, goqu.C("span_count").Desc()},
		{"case7", args{sort: "SPAN_COUNT_ASC"}, goqu.C("span_count").Asc()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chs := &ClickhouseSource{}
			if got := chs.sortConditionStrategy(tt.args.sort); got != tt.want {
				t.Errorf("sortConditionStrategy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClickhouseSource_GetSpans(t *testing.T) {
	chs := ClickhouseSource{
		CompatibleSource: &mockCS{},
		Loader:           &mockLoader{},
		Clickhouse:       &mockclickhouseInf{cli: &mockconn{}},
	}
	spans := chs.GetSpans(context.TODO(), &pb.GetSpansRequest{})
	ass := assert.New(t)
	ass.Equal(1, len(spans))

	chs = ClickhouseSource{
		Loader:     &mockLoader{},
		Clickhouse: &mockclickhouseInf{cli: &mockconn{}},
	}
	spans = chs.GetSpans(context.TODO(), &pb.GetSpansRequest{})
	ass.Equal(0, len(spans))
}

func TestClickhouseSource_GetSpanCount(t *testing.T) {
	chs := &ClickhouseSource{
		Loader: &mockLoader{},
		Clickhouse: &mockclickhouseInf{
			cli: &mockconn{},
		},
	}
	assert.Equal(t, int64(0), chs.GetSpanCount(context.TODO(), "123"))
}

func Test_buildFilter(t *testing.T) {
	type args struct {
		f filter
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{f: filter{
				StartTime: 1652419305504,
				EndTime:   1652508045504,
				OrgName:   "erda",
				TenantID:  "t1",
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE (("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) GROUP BY "trace_id"`,
		},
		{
			name: "traceID",
			args: args{f: filter{
				StartTime: 1652419305504,
				EndTime:   1652508045504,
				OrgName:   "erda",
				TenantID:  "t1",
				Conditions: []custom.Condition{
					{
						TraceId: "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%')) GROUP BY "trace_id"`,
		},
		{
			name: "traceID,duration",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Conditions: []custom.Condition{
					{
						TraceId: "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%')) GROUP BY "trace_id" HAVING (("duration" >= 10000000) AND ("duration" <= 20000000))`,
		},
		{
			name: "traceID,duration,status",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Status:      "trace_all",
				Conditions: []custom.Condition{
					{
						TraceId: "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%')) GROUP BY "trace_id" HAVING (("duration" >= 10000000) AND ("duration" <= 20000000))`,
		},
		{
			name: "traceID,duration,status",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Status:      "trace_error",
				Conditions: []custom.Condition{
					{
						TraceId: "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%') AND (tag_values[indexOf(tag_keys,'error')] = 'true')) GROUP BY "trace_id" HAVING (("duration" >= 10000000) AND ("duration" <= 20000000))`,
		},
		{
			name: "traceID,duration,status",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Status:      "trace_success",
				Conditions: []custom.Condition{
					{
						TraceId: "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%') AND (tag_values[indexOf(tag_keys,'error')] != 'true')) GROUP BY "trace_id" HAVING (("duration" >= 10000000) AND ("duration" <= 20000000))`,
		},
		{
			name: "traceID,duration,status,httpPath",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Status:      "trace_success",
				Conditions: []custom.Condition{
					{
						TraceId:  "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
						HttpPath: "/users",
					},
				},
			}},
			want: "SELECT distinct(trace_id) AS \"trace_id\", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS \"duration\", min(start_time) AS \"min_start_time\" FROM \"spans_all\" WHERE (((\"end_time\" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND (\"org_name\" = 'erda') AND (\"start_time\" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND (\"tenant_id\" = 't1')) AND (\"trace_id\" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%') AND (tag_values[indexOf(tag_keys, 'http_path')] LIKE '%/users%') AND (tag_values[indexOf(tag_keys,'error')] != 'true')) GROUP BY \"trace_id\" HAVING ((\"duration\" >= 10000000) AND (\"duration\" <= 20000000))",
		},
		{
			name: "traceID,duration,status,httpPath,serviceName",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Status:      "trace_success",
				Conditions: []custom.Condition{
					{
						TraceId: "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
					},
					{
						HttpPath: "/users",
					},
					{
						ServiceName: "msp",
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%') AND (tag_values[indexOf(tag_keys, 'http_path')] LIKE '%/users%') AND (tag_values[indexOf(tag_keys, 'service_name')] LIKE '%msp%') AND (tag_values[indexOf(tag_keys,'error')] != 'true')) GROUP BY "trace_id" HAVING (("duration" >= 10000000) AND ("duration" <= 20000000))`,
		},
		{
			name: "traceID,duration,status,httpPath,serviceName,rpcMethod",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Status:      "trace_success",
				Conditions: []custom.Condition{
					{
						TraceId: "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
					},
					{
						HttpPath: "/users",
					},
					{
						ServiceName: "msp",
					},
					{
						RpcMethod: "GetUsers",
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%') AND (tag_values[indexOf(tag_keys, 'http_path')] LIKE '%/users%') AND (tag_values[indexOf(tag_keys, 'service_name')] LIKE '%msp%') AND (tag_values[indexOf(tag_keys, 'rpc_method')] LIKE '%GetUsers%') AND (tag_values[indexOf(tag_keys,'error')] != 'true')) GROUP BY "trace_id" HAVING (("duration" >= 10000000) AND ("duration" <= 20000000))`,
		},
		{
			name: "operator !=",
			args: args{f: filter{
				StartTime:   1652419305504,
				EndTime:     1652508045504,
				OrgName:     "erda",
				TenantID:    "t1",
				DurationMin: 10000000,
				DurationMax: 20000000,
				Status:      "trace_success",
				Conditions: []custom.Condition{
					{
						TraceId:  "972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf",
						Operator: custom.Operator{Operator: "!="},
					},
					{
						HttpPath: "/users",
						Operator: custom.Operator{Operator: "!="},
					},
					{
						ServiceName: "msp",
						Operator:    custom.Operator{Operator: "!="},
					},
					{
						RpcMethod: "GetUsers",
						Operator:  custom.Operator{Operator: "!="},
					},
				},
			}},
			want: `SELECT distinct(trace_id) AS "trace_id", (toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time))) AS "duration", min(start_time) AS "min_start_time" FROM "spans_all" WHERE ((("end_time" <= fromUnixTimestamp64Milli(toInt64(1652508045504))) AND ("org_name" = 'erda') AND ("start_time" >= fromUnixTimestamp64Milli(toInt64(1652419305504))) AND ("tenant_id" = 't1')) AND ("trace_id" NOT LIKE '%972f7ef5-ccc4-4f1a-a0c4-3d60c3dea5cf%') AND (tag_values[indexOf(tag_keys, 'http_path')] NOT LIKE '%/users%') AND (tag_values[indexOf(tag_keys, 'service_name')] NOT LIKE '%msp%') AND (tag_values[indexOf(tag_keys, 'rpc_method')] NOT LIKE '%GetUsers%') AND (tag_values[indexOf(tag_keys,'error')] != 'true')) GROUP BY "trace_id" HAVING (("duration" >= 10000000) AND ("duration" <= 20000000))`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel := goqu.From("spans_all").Select(
				goqu.L("distinct(trace_id)").As("trace_id"),
				goqu.L("(toUnixTimestamp64Nano(max(end_time)) - toUnixTimestamp64Nano(min(start_time)))").As("duration"),
				goqu.L("min(start_time)").As("min_start_time"),
			)
			got := buildFilter(sel, tt.args.f).GroupBy("trace_id")
			sqlstr, _, err := got.ToSQL()
			assert.Nil(t, err)
			assert.Equal(t, tt.want, sqlstr)
		})
	}
}

type mockLoader struct {
}

func (m *mockLoader) ExistsWriteTable(tenant, key string) (ok bool, writeTableName string) {
	panic("implement me")
}

func (m *mockLoader) GetSearchTable(tenant string) (string, *loader.TableMeta) {
	return "spans_all", nil
}

func (m *mockLoader) ReloadTables() chan error {
	panic("implement me")
}

func (m *mockLoader) WatchLoadEvent(listener func(map[string]*loader.TableMeta)) {
	panic("implement me")
}

func (m *mockLoader) WaitAndGetTables(ctx context.Context) map[string]*loader.TableMeta {
	panic("implement me")
}

func (m *mockLoader) Database() string {
	panic("implement me")
}

type mockCS struct {
}

func (m *mockCS) GetSpans(ctx context.Context, req *pb.GetSpansRequest) []*pb.Span {
	return []*pb.Span{{}}
}

func (m *mockCS) GetSpanCount(ctx context.Context, traceID string) int64 {
	return 0
}

func (m *mockCS) GetTraceReqDistribution(ctx context.Context, model custom.Model) ([]*TraceDistributionItem, error) {
	return nil, nil
}

func (m *mockCS) GetTraces(ctx context.Context, req *pb.GetTracesRequest) (*pb.GetTracesResponse, error) {
	return nil, nil
}

type mockclickhouseInf struct {
	cli driver.Conn
}

func (m *mockclickhouseInf) Client() driver.Conn {
	return m.cli
}

func (m *mockclickhouseInf) NewWriter(opts *clickhouse.WriterOptions) *clickhouse.Writer {
	return nil
}

type mockconn struct {
}

func (m *mockconn) Contributors() []string {
	return nil
}

func (m *mockconn) ServerVersion() (*driver.ServerVersion, error) {
	return nil, nil
}

func (m *mockconn) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (m *mockconn) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return &mockRows{}, nil
}

func (m *mockconn) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return &mockRow{}
}

func (m *mockconn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, nil
}

func (m *mockconn) Exec(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

func (m *mockconn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}

func (m *mockconn) Ping(ctx context.Context) error {
	return nil
}

func (m *mockconn) Stats() driver.Stats {
	return driver.Stats{}
}

func (m *mockconn) Close() error {
	return nil
}

type mockRow struct {
}

func (m *mockRow) Err() error {
	panic("implement me")
}

func (m *mockRow) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockRow) ScanStruct(dest interface{}) error {
	panic("implement me")
}

type mockRows struct {
}

func (m *mockRows) Next() bool {
	return false
}

func (m *mockRows) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockRows) ScanStruct(dest interface{}) error {
	return nil
}

func (m *mockRows) ColumnTypes() []driver.ColumnType {
	return nil
}

func (m *mockRows) Totals(dest ...interface{}) error {
	return nil
}

func (m *mockRows) Columns() []string {
	return nil
}

func (m *mockRows) Close() error {
	return nil
}

func (m *mockRows) Err() error {
	return nil
}
