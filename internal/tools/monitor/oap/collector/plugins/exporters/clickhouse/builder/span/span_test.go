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

package span

import (
	"context"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/plugins/exporters/clickhouse/builder"
)

func TestBuilder_buildBatches(t *testing.T) {
	type args struct {
		items []*trace.Span
	}
	tests := []struct {
		name    string
		args    args
		want    []driver.Batch
		wantErr bool
	}{
		{
			args: args{
				items: []*trace.Span{
					{
						StartTime:     1652421595810000000,
						EndTime:       1652421595810843400,
						OrgName:       "erda",
						TraceId:       "bac6e329-4be5-4ed5-b364-c0fc305b4f8e",
						SpanId:        "4e4a3048-7ac0-4233-a133-c63cb3293e39",
						ParentSpanId:  "411d3049-0053-44c4-a96b-997e277fd08e",
						OperationName: "Mysql/PreparedStatement/executeUpdate",
						Tags: map[string]string{
							"application_id":   "8931",
							"application_name": "trantor-datastore",
							"cluster_name":     "terminus-captain",
							"component":        "Mysql",
							"db_host":          "mysql-master.addon-mysql--wb8019b0757474e1c88b5dcaa64a5ab1f.svc.cluster.local:3306",
							"db_instance":      "trantor_console_staging",
							"db_statement":     "update `meta_store_management__versioned_model_field` set `FromModule` = ? where `id` in (?)",
							"db_system":        "Mysql",
							"env_id":           "c75de7278874f37d9f1bc818b473fc23",
							"host":             "node-010167000123",
							"host_ip":          "10.167.0.123",
							"org_id":           "2",
							"org_name":         "terminus",
							"terminus_key":     "c75de7278874f37d9f1bc818b473fc23",
							"workspace":        "STAGING",
						},
					},
				},
			},
			want: []driver.Batch{
				&mockBatch{arr: []trace.TableSpan{
					{
						StartTime:     time.Unix(0, 1652421595810000000),
						EndTime:       time.Unix(0, 1652421595810843400),
						OrgName:       "erda",
						TenantId:      "c75de7278874f37d9f1bc818b473fc23",
						TraceId:       "bac6e329-4be5-4ed5-b364-c0fc305b4f8e",
						SpanId:        "4e4a3048-7ac0-4233-a133-c63cb3293e39",
						ParentSpanId:  "411d3049-0053-44c4-a96b-997e277fd08e",
						OperationName: "Mysql/PreparedStatement/executeUpdate",
						TagKeys:       []string{"application_id", "application_name", "cluster_name", "component", "db_host", "db_instance", "db_statement", "db_system", "env_id", "host", "host_ip", "org_id", "org_name", "terminus_key", "workspace"},
						TagValues:     []string{"8931", "trantor-datastore", "terminus-captain", "Mysql", "mysql-master.addon-mysql--wb8019b0757474e1c88b5dcaa64a5ab1f.svc.cluster.local:3306", "trantor_console_staging", "update `meta_store_management__versioned_model_field` set `FromModule` = ? where `id` in (?)", "Mysql", "c75de7278874f37d9f1bc818b473fc23", "node-010167000123", "10.167.0.123", "2", "terminus", "c75de7278874f37d9f1bc818b473fc23", "STAGING"},
					},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bu := &Builder{
				cfg:     &builder.BuilderConfig{TenantIdKey: "terminus_key"},
				Creator: &mockCreator{},
				client:  &mockConn{},
			}
			got, err := bu.buildBatches(context.TODO(), tt.args.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildBatches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
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
	arr []trace.TableSpan
}

func (m *mockBatch) Abort() error { return nil }

func (m *mockBatch) Append(v ...interface{}) error { return nil }

func (m *mockBatch) AppendStruct(v interface{}) error {
	m.arr = append(m.arr, *v.(*trace.TableSpan))
	return nil
}

func (m *mockBatch) Column(i int) driver.BatchColumn { return nil }

func (m *mockBatch) Send() error { return nil }

func (m *mockBatch) Flush() error { return nil }

func (m *mockBatch) IsSent() bool { return true }

func (m *mockBatch) Rows() int { return 0 }
