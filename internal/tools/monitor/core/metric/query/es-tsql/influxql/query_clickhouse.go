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
	"io"
	"reflect"
	"time"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

type QueryClickhouse struct {
	start, end int64
	debug      bool
	kind       string

	subLiters map[string]string
	sources   []*model.Source
	ctx       *Context
	expr      *goqu.SelectDataset
	column    []*SQLColumnHandler

	orgName     string
	terminusKey string
}

func (q QueryClickhouse) Sources() []*model.Source {
	return q.sources
}

func (q QueryClickhouse) SearchSource() interface{} {
	return q.expr
}

func (q QueryClickhouse) AppendBoolFilter(key string, value interface{}) {
	q.expr = q.expr.Where(goqu.L(key).Eq(value))
}

var getData func(row ckdriver.Rows) (map[string]interface{}, error)

func (q QueryClickhouse) ParseResult(resp interface{}) (*model.Data, error) {
	if resp == nil {
		return nil, nil
	}

	rows, ok := resp.(ckdriver.Rows)
	rs := &model.Data{}
	if !ok {
		return nil, errors.New("data should be ck driver.Rows")
	}

	var (
		columnTypes = rows.ColumnTypes()
		vars        = make([]interface{}, len(columnTypes))
		columns     = rows.Columns()
		err         error
	)

	for i := range columnTypes {
		vars[i] = reflect.New(columnTypes[i].ScanType()).Interface()
	}

	for _, c := range q.column {
		if c.col == nil {
			c.col = &model.Column{
				Name: getColumnName(c.field),
			}
			rs.Columns = append(rs.Columns, c.col)
		}
	}

	getData = func(row ckdriver.Rows) (map[string]interface{}, error) {
		if err := rows.Scan(vars...); err != nil {
			return nil, err
		}

		data := make(map[string]interface{})
		for i, v := range vars {
			columnName := columns[i]
			data[columnName] = pretty(v)
		}
		return data, nil
	}

	cur := make(map[string]interface{})
	next := make(map[string]interface{})

	q.ctx.AttributesCache()

	// first read
	rows.Next()
	cur, err = getData(rows)
	isTail := false
	if err != nil {
		if err == io.EOF {
			// The first item is empty
			return rs, nil
		}
		return nil, err
	}

	for {
		if cur == nil && next == nil {
			break
		}

		if isTail {
			break
		}

		if len(cur) <= 0 {
			cur, err = getData(rows)
			if err != nil {
				return nil, err
			}
		}

		if rows.Next() {
			next, err = getData(rows)
			q.ctx.attributesCache["next"] = next
		} else {
			isTail = true
		}

		var row []interface{}
		for _, c := range q.column {
			v, err := c.getValue(cur)
			if err != nil {
				return nil, err
			}
			row = append(row, v)
		}
		rs.Rows = append(rs.Rows, row)

		cur = next
		next = nil
	}
	rs.Total = int64(len(rs.Rows))
	rs.Interval = q.ctx.Interval()
	return rs, nil
}

func pretty(data interface{}) interface{} {
	if data == nil {
		return ""
	}
	switch v := data.(type) {
	case *float64:
		return *v
	case *float32:
		return *v
	case *int32:
		return *v
	case *int64:
		return *v
	case *int:
		return *v
	case *uint:
		return *v
	case *uint32:
		return *v
	case *uint64:
		return *v
	case *string:
		return *v
	case *time.Time:
		return v.UnixNano()
	}

	return data
}

func (q QueryClickhouse) Context() tsql.Context {
	return q.ctx
}

func (q QueryClickhouse) Debug() bool {
	return q.debug
}

func (q QueryClickhouse) Timestamp() (int64, int64) {
	return q.start, q.end
}

func (q QueryClickhouse) Kind() string {
	return q.kind
}

func (q QueryClickhouse) SubSearchSource() interface{} {
	return q.subLiters
}

func (q QueryClickhouse) OrgName() string {
	return q.orgName
}
func (q QueryClickhouse) TerminusKey() string {
	return q.terminusKey
}
