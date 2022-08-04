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
	"context"
	"reflect"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
)

type mockClickhouse struct {
	mchRow *mockClickhouseRow
	sql    string
}

type mockClickhouseRow struct {
	column     []string
	columnType []interface{}
	err        error
	data       [][]interface{}
	point      int
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

func Test(t *testing.T) {
	q := QueryClickhouse{
		column: []*SQLColumnHandler{
			{
				col: &model.Column{
					Name: "service_id",
				},
			},
		},
		ctx: &Context{},
	}
	row := mockClickhouseRow{
		column: []string{"service_id"},
		columnType: []interface{}{
			reflect.String,
		},
		data: [][]interface{}{
			{
				"111",
			},
			{
				"222",
			},
			{
				"333",
			},
		},
	}
	data, err := parse(q, &row)
	require.NoError(t, err)
	require.Equal(t, 3, len(data.Rows))
	require.Equal(t, 1, len(data.Columns))

}

func parse(q QueryClickhouse, rows driver.Rows) (*model.Data, error) {
	return q.ParseResult(context.Background(), rows)
}
