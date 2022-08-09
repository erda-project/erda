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
	"io"
	"reflect"
	"strings"
	"time"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
	"github.com/erda-project/erda/pkg/common/trace"
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

func iterate(ctx context.Context, row ckdriver.Rows) (map[string]interface{}, error) {
	_, span := otel.Tracer("parser").Start(ctx, "parser.iterate")
	defer span.End()

	var (
		columnTypes = row.ColumnTypes()
		vars        = make([]interface{}, len(columnTypes))
		columns     = row.Columns()
	)

	span.SetAttributes(attribute.String("origin.columns", strings.Join(columns, ";")))

	for i := range columnTypes {
		vars[i] = reflect.New(columnTypes[i].ScanType()).Interface()
	}

	if row.Err() != nil {
		return nil, errors.Wrap(row.Err(), "failed to get data")
	}

	if err := row.Scan(vars...); err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	for i, v := range vars {
		if v == nil {
			continue
		}
		columnName := columns[i]
		data[columnName] = pretty(v)
	}

	span.AddEvent("dump", oteltrace.WithAttributes(trace.BigStringAttribute("data", data)))
	return data, nil
}
func (q QueryClickhouse) ParseResult(ctx context.Context, resp interface{}) (*model.Data, error) {
	newCtx, span := otel.Tracer("parser").Start(ctx, "parser.result")
	defer span.End()

	if resp == nil {
		return nil, nil
	}

	var err error

	rows, ok := resp.(ckdriver.Rows)
	rs := &model.Data{}
	if !ok {
		return nil, errors.New("data should be ck driver.Rows")
	}

	for _, c := range q.column {
		if c.col == nil {
			c.col = &model.Column{
				Name: getColumnName(c.field),
			}
			// delete *
			if c.AllColumns() {
				continue
			}
			rs.Columns = append(rs.Columns, c.col)
		}
	}

	cur := make(map[string]interface{})
	next := make(map[string]interface{})

	q.ctx.AttributesCache()

	// first read
	rows.Next()
	cur, err = iterate(newCtx, rows)
	isTail := false
	if err != nil {
		if err == io.EOF {
			// The first item is empty
			return rs, nil
		}

		span.RecordError(err, oteltrace.WithAttributes(attribute.String("comment", "first is error")))
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
			cur, err = iterate(newCtx, rows)
			if err != nil {
				span.RecordError(err)
				return nil, err
			}
		}

		if rows.Next() {
			next, err = iterate(newCtx, rows)
			if err != nil {
				span.RecordError(err)
				return nil, err
			}
			q.ctx.attributesCache["next"] = next
		} else {
			isTail = true
		}

		var row []interface{}
		for _, c := range q.column {
			if c.AllColumns() {
				allValue, err := c.getALLValue(cur)
				if err != nil {
					span.RecordError(err, oteltrace.WithAttributes(attribute.String("comment", "get all value error")))
					return nil, err
				}
				for k, v := range allValue {
					var point int
					rs.Columns, point = appendIfMissing(rs.Columns, &model.Column{
						Name: k,
					})
					if cap(row) <= point || row == nil {
						newRow := make([]interface{}, point+1)
						copy(newRow, row)
						row = newRow
					}
					row[point] = v
				}
				continue
			}

			v, err := c.getValue(cur)
			if err != nil {
				span.RecordError(err, oteltrace.WithAttributes(
					attribute.String("comment", "get column value error"),
					trace.BigStringAttribute("data", cur),
				))
				return nil, err
			}

			var point int
			rs.Columns, point = appendIfMissing(rs.Columns, c.col)
			if cap(row) <= point || row == nil {
				newRow := make([]interface{}, point+1)
				copy(newRow, row)
				row = newRow
			}
			row[point] = v
		}
		if row != nil {
			rs.Rows = append(rs.Rows, row)
		}

		cur = next
		next = nil
	}
	rs.Total = int64(len(rs.Rows))
	rs.Interval = q.ctx.Interval()
	span.SetAttributes(trace.BigStringAttribute("result.total", len(rs.Rows)))
	return rs, nil
}

func appendIfMissing(slice []*model.Column, target *model.Column) ([]*model.Column, int) {
	for i := 0; i < len(slice); i++ {
		if slice[i].Name == target.Name {
			return slice, i
		}
	}
	return append(slice, target), len(slice)
}
func pretty(data interface{}) interface{} {
	if data == nil {
		return data
	}
	switch v := data.(type) {
	case *float64:
		return *v
	case **float64:
		if *v == nil {
			return float64(0)
		}
		return **v
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
	case *[]string:
		return *v
	case *[]float64:
		return *v
	default:
		return v
	}
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
