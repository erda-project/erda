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
		want string
	}{
		{"case1", args{sort: ""}, "ORDER BY min_start_time DESC"},
		{"case2", args{sort: "TRACE_TIME_DESC"}, "ORDER BY min_start_time DESC"},
		{"case3", args{sort: "TRACE_TIME_ASC"}, "ORDER BY min_start_time ASC"},
		{"case4", args{sort: "TRACE_DURATION_DESC"}, "ORDER BY duration DESC"},
		{"case5", args{sort: "TRACE_DURATION_ASC"}, "ORDER BY duration ASC"},
		{"case6", args{sort: "SPAN_COUNT_DESC"}, "ORDER BY span_count DESC"},
		{"case7", args{sort: "SPAN_COUNT_ASC"}, "ORDER BY span_count ASC"},
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

func TestClickhouseSource_sortConditionStrategy1(t *testing.T) {
	type args struct {
		sort string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{sort: "span_count_desc"},
			want: "ORDER BY span_count DESC",
		},
		{
			args: args{sort: "span_count_asc"},
			want: "ORDER BY span_count ASC",
		},
		{
			args: args{sort: "trace_duration_desc"},
			want: "ORDER BY duration DESC",
		},
		{
			args: args{sort: "trace_duration_asc"},
			want: "ORDER BY duration ASC",
		},
		{
			args: args{sort: "trace_time_desc"},
			want: "ORDER BY min_start_time DESC",
		},
		{
			args: args{sort: "trace_time_asc"},
			want: "ORDER BY min_start_time ASC",
		},
		{
			args: args{sort: "xxx"},
			want: "ORDER BY min_start_time DESC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chs := &ClickhouseSource{}
			assert.Equal(t, tt.want, chs.sortConditionStrategy(tt.args.sort))
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

func (m mockclickhouseInf) NewWriter(opts *clickhouse.WriterOptions) *clickhouse.Writer {
	return nil
}

type mockconn struct {
}

func (m mockconn) Contributors() []string {
	return nil
}

func (m mockconn) ServerVersion() (*driver.ServerVersion, error) {
	return nil, nil
}

func (m mockconn) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

type mockRows struct {
}

func (m mockRows) Next() bool {
	return false
}

func (m mockRows) Scan(dest ...interface{}) error {
	return nil
}

func (m mockRows) ScanStruct(dest interface{}) error {
	return nil
}

func (m mockRows) ColumnTypes() []driver.ColumnType {
	return nil
}

func (m mockRows) Totals(dest ...interface{}) error {
	return nil
}

func (m mockRows) Columns() []string {
	return nil
}

func (m mockRows) Close() error {
	return nil
}

func (m mockRows) Err() error {
	return nil
}

func (m mockconn) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return &mockRows{}, nil
}

func (m mockconn) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return nil
}

func (m mockconn) PrepareBatch(ctx context.Context, query string) (driver.Batch, error) {
	return nil, nil
}

func (m mockconn) Exec(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

func (m mockconn) AsyncInsert(ctx context.Context, query string, wait bool) error {
	return nil
}

func (m mockconn) Ping(ctx context.Context) error {
	return nil
}

func (m mockconn) Stats() driver.Stats {
	return driver.Stats{}
}

func (m mockconn) Close() error {
	return nil
}
