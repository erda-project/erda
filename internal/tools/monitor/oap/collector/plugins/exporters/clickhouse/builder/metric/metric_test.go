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

package metric

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/settings/retention-strategy"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
)

func TestBuilder_buildBatches(t *testing.T) {
	type args struct {
		items []*metric.Metric
	}
	tests := []struct {
		name    string
		args    args
		want    []driver.Batch
		wantErr bool
	}{
		{
			args: args{items: []*metric.Metric{
				{
					Name:      "cpu",
					Timestamp: 1652421595810000000,
					Tags: map[string]string{
						"cluster_name": "dev",
					},
					Fields: map[string]interface{}{
						"usage": 0.1,
					},
					OrgName: "erda",
				},
			}},
			want: []driver.Batch{
				&mockBatch{arr: []interface{}{
					&metric.TableMetrics{
						OrgName:           "erda",
						MetricGroup:       "cpu",
						Timestamp:         time.Unix(0, 1652421595810000000),
						NumberFieldKeys:   []string{"usage"},
						NumberFieldValues: []float64{0.1},
						StringFieldKeys:   []string{},
						StringFieldValues: []string{},
						TagKeys:           []string{"cluster_name"},
						TagValues:         []string{"dev"},
					},
				}},
				&mockBatch{arr: []interface{}{
					&metric.TableMetricsMeta{
						OrgName:         "erda",
						MetricGroup:     "cpu",
						Timestamp:       time.Unix(0, 1652421595810000000),
						NumberFieldKeys: []string{"usage"},
						StringFieldKeys: []string{},
						TagKeys:         []string{"cluster_name"},
					}},
				}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bu := &Builder{
				cfg:       &builder.BuilderConfig{TenantIdKey: "terminus_key"},
				Creator:   &mockCreator{},
				client:    &mockConn{},
				Loader:    &mockLoader{},
				Retention: &mockRetention{},
			}
			got, err := bu.buildBatches(context.TODO(), tt.args.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildBatches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildBatches() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockCreator struct {
}

func (m *mockCreator) Ensure(ctx context.Context, tenant, key string, ttlDays int64) (<-chan error, string) {
	return nil, trace.CH_TABLE
}

type mockConn struct {
}

func (m *mockConn) Contributors() []string {
	panic("implement me")
}

func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) {
	panic("implement me")
}

func (m *mockConn) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	panic("implement me")
}

func (m *mockConn) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	panic("implement me")
}

func (m *mockConn) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	panic("implement me")
}

func (m *mockConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return &mockBatch{}, nil
}

func (m *mockConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	panic("implement me")
}

func (m *mockConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	panic("implement me")
}

func (m *mockConn) Ping(ctx context.Context) error {
	panic("implement me")
}

func (m *mockConn) Stats() driver.Stats {
	panic("implement me")
}

func (m *mockConn) Close() error {
	panic("implement me")
}

type mockBatch struct {
	arr []interface{}
}

func (m *mockBatch) Abort() error { return nil }

func (m *mockBatch) Append(v ...interface{}) error { return nil }

func (m *mockBatch) AppendStruct(v interface{}) error {
	m.arr = append(m.arr, v)
	return nil
}

func (m *mockBatch) Column(i int) driver.BatchColumn { return nil }

func (m *mockBatch) Send() error { return nil }

func (m *mockBatch) Flush() error { return nil }

func (m *mockBatch) IsSent() bool { return true }

func (m *mockBatch) Rows() int { return 0 }

type mockLoader struct {
}

func (m *mockLoader) ExistsWriteTable(tenant, key string) (ok bool, writeTableName string) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLoader) GetSearchTable(tenant string) (string, *loader.TableMeta) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLoader) ReloadTables() chan error {
	//TODO implement me
	panic("implement me")
}

func (m *mockLoader) WatchLoadEvent(listener func(map[string]*loader.TableMeta)) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLoader) WaitAndGetTables(ctx context.Context) map[string]*loader.TableMeta {
	//TODO implement me
	panic("implement me")
}

func (m *mockLoader) Database() string {
	return "test"
}

type mockRetention struct {
}

func (m mockRetention) Default() *retention.TTL {
	//TODO implement me
	panic("implement me")
}

func (m mockRetention) DefaultHotDataTTL() time.Duration {
	//TODO implement me
	panic("implement me")
}

func (m mockRetention) GetTTL(key string) *retention.TTL {
	return &retention.TTL{}
}

func (m mockRetention) DefaultTTL() time.Duration {
	//TODO implement me
	panic("implement me")
}

func (m mockRetention) GetConfigKey(name string, tags map[string]string) string {
	return ""
}

func (m mockRetention) GetTTLByTags(name string, tags map[string]string) time.Duration {
	//TODO implement me
	panic("implement me")
}

func (m mockRetention) Loading(ctx context.Context) {
	//TODO implement me
	panic("implement me")
}
