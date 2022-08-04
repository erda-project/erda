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

package esinfluxql

import (
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
)

type mockClickhouse struct {
	mchRow *mockClickhouseRow
	sql    string
}

type mockClickhouseRow struct {
	column []string
	err    error
	data   [][]interface{}
	point  int
}

func (m *mockClickhouseRow) Next() bool {
	if m.point < len(m.data) {
		m.point = m.point + 1
		return true
	}
	return false
}

func (m *mockClickhouseRow) Scan(dest ...interface{}) error {
	return nil
}

func (m *mockClickhouseRow) ScanStruct(dest interface{}) error {
	return nil
}

func (m *mockClickhouseRow) ColumnTypes() []driver.ColumnType {
	return nil
}

func (m *mockClickhouseRow) Totals(dest ...interface{}) error {
	return nil
}

func (m *mockClickhouseRow) Columns() []string {
	return m.column
}

func (m mockClickhouseRow) Close() error {
	return nil
}

func (m *mockClickhouseRow) Err() error {
	return m.err
}

func (m *mockClickhouse) QueryRaw(orgName string, expr *goqu.SelectDataset) (driver.Rows, error) {
	if expr != nil {
		m.sql, _, _ = expr.ToSQL()
	}
	return m.mchRow, nil
}

func TestMock(t *testing.T) {
	// TODO

}
