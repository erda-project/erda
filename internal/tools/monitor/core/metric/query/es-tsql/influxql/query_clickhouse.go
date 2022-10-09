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
	"math"
	"reflect"
	"strings"
	"time"

	ckdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/doug-martin/goqu/v9"
	"github.com/influxdata/influxql"
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
	var (
		columnTypes = row.ColumnTypes()
		vars        = make([]interface{}, len(columnTypes))
		columns     = row.Columns()
	)

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

	return data, nil
}
func (q QueryClickhouse) ParseResult(ctx context.Context, resp interface{}) (*model.Data, error) {
	newCtx, span := otel.Tracer("parser").Start(ctx, "parser.result")
	defer span.End()

	if resp == nil {
		return nil, nil
	}

	rows, ok := resp.(ckdriver.Rows)
	rs := &model.Data{}
	if !ok {
		return nil, errors.New("data should be ck driver.Rows")
	}

	q.buildColumn(newCtx, rs)

	err := q.parse(newCtx, rows, rs)
	if err != nil {
		return nil, errors.Wrap(err, "clickhouse metric query parse is fail")
	}
	rs.Total = int64(len(rs.Rows))
	rs.Interval = q.ctx.Interval()
	span.SetAttributes(trace.BigStringAttribute("result.total", len(rs.Rows)))
	return rs, nil
}

func (q QueryClickhouse) fetchAllValue(ctx context.Context, rows ckdriver.Rows) ([]map[string]interface{}, error) {
	newCtx, span := otel.Tracer("parser").Start(ctx, "fetch")
	defer span.End()
	var dates []map[string]interface{}
	for nextRows(newCtx, rows) {
		cur, err := iterate(newCtx, rows)
		if err != nil {
			return nil, err
		}
		dates = append(dates, cur)
	}
	return dates, nil
}

func (q QueryClickhouse) parse(ctx context.Context, rows ckdriver.Rows, rs *model.Data) error {
	newCtx, span := otel.Tracer("parser").Start(ctx, "parse")
	defer span.End()

	var err error
	q.ctx.AttributesCache()

	dates, err := q.fetchAllValue(newCtx, rows)
	if err != nil {
		return errors.Wrap(err, "fetch clickhouse value is failed")
	}

	if len(dates) <= 0 {
		return nil
	}

	step := getStep(newCtx, dates)

	for i, iterate := range dates {
		var row []interface{}
		point := i + step
		if point < len(dates) {
			q.ctx.attributesCache["next"] = dates[i+step]
		}

		for _, c := range q.column {
			if c.AllColumns() {
				row, err = q.parseAllColumn(c, iterate, rs)
				if err != nil {
					return err
				}
				continue
			}

			v, err := c.getValue(newCtx, iterate)
			if err != nil {
				span.RecordError(err, oteltrace.WithAttributes(
					attribute.String("comment", "get column value error"),
					trace.BigStringAttribute("data", iterate),
				))
				return err
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
	}
	return nil
}

// rows must be sorted
func getStep(ctx context.Context, rows []map[string]interface{}) int {
	_, span := otel.Tracer("parser").Start(ctx, "calculator.step")
	defer span.End()

	if len(rows) <= 0 {
		return 0
	}
	timestamp := int64(math.MaxInt64)
	i := 0
	for _, row := range rows {
		_t, exist := row["bucket_timestamp"]
		if !exist {
			break
		}
		_timestamp, ok := _t.(int64)
		if !ok {
			break
		}

		if _timestamp > timestamp {
			span.SetAttributes(attribute.Int("step", i))
			return i
		} else {
			timestamp = _timestamp
		}
		i++
	}
	return 0
}

func (q QueryClickhouse) parseAllColumn(c *SQLColumnHandler, cur map[string]interface{}, rs *model.Data) ([]interface{}, error) {
	allValue, err := c.getALLValue(cur)
	var row []interface{}
	if err != nil {
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
	return row, nil
}
func (q QueryClickhouse) buildColumn(ctx context.Context, rs *model.Data) {
	_, span := otel.Tracer("parser").Start(ctx, "build.column")
	defer span.End()
	for _, c := range q.column {
		if c.col == nil {
			c.col = &model.Column{
				Name: getColumnName(c.field),
			}
			key, flag := getExprStringAndFlag(c.field.Expr, influxql.AnyField)
			c.col.Flag = flag
			if q.ctx.dimensions[key] {
				c.col.Flag |= model.ColumnFlagGroupBy
			}

			// delete *
			if c.AllColumns() {
				continue
			}
			rs.Columns = append(rs.Columns, c.col)
		}
	}
}

func nextRows(ctx context.Context, rows ckdriver.Rows) bool {
	return rows.Next()
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
			return nil
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
	case **uint64:
		if *v == nil {
			return nil
		}
		return **v
	case *string:
		return *v
	case **string:
		if *v == nil {
			return ""
		}
		return **v
	case *time.Time:
		return v.UnixNano()
	case **time.Time:
		return (*v).UnixNano()
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

func (q QueryClickhouse) OrgName() []string {
	if strings.Index(q.orgName, ",") != -1 {
		return strings.Split(q.orgName, ",")
	}
	return []string{q.orgName}
}
func (q QueryClickhouse) TerminusKey() string {
	return q.terminusKey
}
