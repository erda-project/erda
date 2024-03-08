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

package meta

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/clickhouse"
)

type mockClickhouse struct {
	mockResult *mockResult
	verify     func(sql string)
}

func (m mockClickhouse) Client() driver.Conn {
	return &mockClickhouseConn{
		mockResult: m.mockResult,
		verify:     m.verify,
	}
}

func (*mockClickhouse) NewWriter(opts *clickhouse.WriterOptions) *clickhouse.Writer {
	return nil
}

type mockClickhouseConn struct {
	verify     func(sql string)
	mockResult *mockResult
}

func (m *mockClickhouseConn) Contributors() []string {
	return []string{}
}

func (m *mockClickhouseConn) ServerVersion() (*driver.ServerVersion, error) {
	return nil, nil
}

func (m *mockClickhouseConn) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (m *mockClickhouseConn) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	if m.verify != nil {
		m.verify(query)
	}
	return m.mockResult, nil
}

func (m *mockClickhouseConn) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return nil
}

func (m *mockClickhouseConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, nil
}

func (m *mockClickhouseConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

func (m *mockClickhouseConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}

func (m *mockClickhouseConn) Ping(ctx context.Context) error {
	return nil
}

func (m *mockClickhouseConn) Stats() driver.Stats {
	return driver.Stats{}
}

func (m *mockClickhouseConn) Close() error {
	return nil
}

type mockResult struct {
	data  []MetricMeta
	point int
}

func (m *mockResult) Next() bool {
	if len(m.data) <= 0 {
		return false
	}
	if m.point < len(m.data) {
		m.point++
		return true
	}
	return false
}

func (m *mockResult) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockResult) ScanStruct(dest interface{}) error {
	data := m.data[m.point-1]

	v, ok := dest.(*MetricMeta)
	if !ok {
		return errors.New("error type")
	}
	v.MetricGroup = data.MetricGroup
	v.NumberKeys = data.NumberKeys
	v.TagKeys = data.TagKeys
	v.StringKeys = data.StringKeys
	v.Scope = data.Scope
	v.ScopeId = data.ScopeId
	return nil
}

func (m *mockResult) ColumnTypes() []driver.ColumnType {
	return nil
}

func (m *mockResult) Totals(dest ...interface{}) error {
	return nil
}

func (m *mockResult) Columns() []string {
	return nil
}

func (m *mockResult) Close() error {
	return nil
}

func (m *mockResult) Err() error {
	return nil
}
