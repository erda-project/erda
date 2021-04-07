// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package esinfluxql

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
	"github.com/influxdata/influxql"
	"github.com/olivere/elastic"
)

// Query .
type Query struct {
	sources      []*tsql.Source
	searchSource *elastic.SearchSource
	boolQuery    *elastic.BoolQuery
	stmt         *influxql.SelectStatement
	columns      []*columnHandler
	flag         queryFlag
	aggs         map[string]elastic.Aggregation
	ctx          *Context
	allColumnsFn func(start, end int64, sources []*tsql.Source) ([]*tsql.Column, error)
}

type queryFlag int32

// queryFlag .
const (
	queryFlagNone    = queryFlag(0)
	queryFlagColumns = queryFlag(1 << (iota - 1))
	queryFlagAllColumns
	queryFlagDimensions
	queryFlagAggs
	queryFlagGroupByTime
	queryFlagGroupByRange
)

// Columns .
type Columns []*tsql.Column

func (cs Columns) Len() int { return len(cs) }
func (cs Columns) Less(i, j int) bool {
	if cs[i].Key == cs[j].Key {
		return cs[i].Name < cs[j].Name
	}
	return cs[i].Key < cs[j].Key
}
func (cs Columns) Swap(i, j int) { cs[i], cs[j] = cs[j], cs[i] }

// Sources .
func (q *Query) Sources() []*tsql.Source { return q.sources }

// SearchSource .
func (q *Query) SearchSource() *elastic.SearchSource { return q.searchSource }

// BoolQuery .
func (q *Query) BoolQuery() *elastic.BoolQuery { return q.boolQuery }

// SetAllColumnsCallback .
func (q *Query) SetAllColumnsCallback(fn func(start, end int64, sources []*tsql.Source) ([]*tsql.Column, error)) {
	q.allColumnsFn = fn
}

// Context .
func (q *Query) Context() tsql.Context { return q.ctx }

// ParseResult .
func (q *Query) ParseResult(resp *elastic.SearchResult) (*tsql.ResultSet, error) {
	if resp != nil {
		q.ctx.aggregations = resp.Aggregations
	}
	rs := &tsql.ResultSet{
		Interval: q.ctx.Interval(),
	}
	if resp != nil {
		rs.Total = resp.TotalHits()
	}
	if q.flag&(queryFlagDimensions|queryFlagAggs) != queryFlagNone {
		return q.parseAggData(resp, rs)
	}
	return q.parseRawData(resp, rs)
}

func (q *Query) parseRawData(resp *elastic.SearchResult, rs *tsql.ResultSet) (*tsql.ResultSet, error) {
	if q.flag == queryFlagNone {
		for _, c := range q.columns {
			if c.col == nil {
				c.col = &tsql.Column{
					Name: getColumnName(c.field),
				}
			}
			rs.Columns = append(rs.Columns, c.col)
		}
		var values []interface{}
		source := make(map[string]interface{})
		for _, c := range q.columns {
			v, err := c.getRawValue(source)
			if err != nil {
				return nil, err
			}
			values = append(values, v)
		}
		rs.Rows = append(rs.Rows, values)
		return rs, nil
	}
	var sources []map[string]interface{}
	if resp != nil && resp.Hits != nil {
		for _, hit := range resp.Hits.Hits {
			var source map[string]interface{}
			err := json.Unmarshal([]byte(*hit.Source), &source)
			if err != nil {
				continue
			}
			sources = append(sources, source)
		}
	}
	var allColumns bool
	for _, c := range q.columns {
		if c.AllColumns() {
			allColumns = true
			break
		}
	}
	var columns Columns
	if allColumns {
		if q.allColumnsFn != nil {
			cols, err := q.allColumnsFn(q.ctx.start, q.ctx.end, q.sources)
			if err != nil {
				return nil, err
			}
			columns = cols
		}
		if len(columns) <= 0 && len(sources) > 0 {
			cols := make(map[string]string)
			for _, source := range sources {
				parseSourceColumns("", source, cols)
			}
			for c := range cols {
				if strings.HasPrefix(c, "@") || strings.HasPrefix(c, "tags._") {
					continue
				}
				col := &tsql.Column{
					Key:  c,
					Name: c,
				}
				if strings.HasPrefix(c, tsql.FieldsKey) {
					col.Flag = tsql.ColumnFlagField
					col.Name = c[len(tsql.FieldsKey):] + "::field"
				} else if strings.HasPrefix(c, tsql.TagsKey) {
					col.Flag = tsql.ColumnFlagTag
					col.Name = c[len(tsql.TagsKey):] + "::tag"
				} else {
					if c == tsql.NameKey {
						col.Flag = tsql.ColumnFlagName
					} else if c == q.ctx.TimeKey() {
						col.Flag = tsql.ColumnFlagTimestamp
					}
				}
				columns = append(columns, col)
			}
		}
		sort.Sort(columns)
	}
	for _, c := range q.columns {
		if c.AllColumns() {
			rs.Columns = append(rs.Columns, columns...)
		} else {
			if c.col == nil {
				c.col = &tsql.Column{
					Name: getColumnName(c.field),
				}
			}
			rs.Columns = append(rs.Columns, c.col)
		}
	}
	for _, source := range sources {
		var values []interface{}
		for _, c := range q.columns {
			if c.AllColumns() {
				for _, col := range columns {
					v := getGetValueFromFlatMap(source, col.Key, ".")
					values = append(values, v)
				}
			} else {
				v, err := c.getRawValue(source)
				if err != nil {
					return nil, err
				}
				values = append(values, v)
			}
		}
		q.ctx.row++
		rs.Rows = append(rs.Rows, values)
	}
	return rs, nil
}

func (q *Query) parseAggData(resp *elastic.SearchResult, rs *tsql.ResultSet) (*tsql.ResultSet, error) {
	for _, c := range q.columns {
		if c.AllColumns() {
			return nil, fmt.Errorf("not support field * if has aggregation function or group by")
		}
		if c.col == nil {
			c.col = &tsql.Column{
				Name: getColumnName(c.field),
			}
			key, flag := getExprStringAndFlag(c.field.Expr, influxql.AnyField)
			c.col.Flag = flag
			if q.ctx.dimensions[key] {
				c.col.Flag |= tsql.ColumnFlagGroupBy
			}
		}
		rs.Columns = append(rs.Columns, c.col)
	}
	if resp == nil {
		return rs, nil
	}
	err := q.parseDimensionsAggsData(rs, resp.Aggregations, nil)
	if err != nil {
		return nil, err
	}
	// if q.flag&(queryFlagDimensions|queryFlagAggs) != queryFlagNone {
	// 	SortResultSet(rs, q.stmt.SortFields)
	// 	if len(rs.Rows) < q.stmt.Offset {
	// 		rs.Rows = rs.Rows[0:0]
	// 	} else {
	// 		rs.Rows = rs.Rows[q.stmt.Offset:]
	// 	}
	// 	if len(rs.Rows) > q.stmt.Limit {
	// 		rs.Rows = rs.Rows[0:q.stmt.Limit]
	// 	}
	// }
	return rs, nil
}

func (q *Query) parseDimensionsAggsData(rs *tsql.ResultSet, aggs elastic.Aggregations, buckets []interface{}) error {
	if terms, ok := aggs.Terms("term"); ok {
		if len(terms.Buckets) > q.stmt.Offset {
			// 有分组的情况下，offset的作用是跳过多少分组
			for _, bucket := range terms.Buckets[q.stmt.Offset:] {
				err := q.parseDimensionsAggsData(rs, bucket.Aggregations, append(buckets, bucket))
				if err != nil {
					return err
				}
			}
		}
	} else if histogram, ok := aggs.Histogram("histogram"); ok {
		var list []*elastic.AggregationBucketHistogramItem
		for i := len(histogram.Buckets) - 1; i >= 0; i-- {
			if histogram.Buckets[i].DocCount > 0 {
				list = histogram.Buckets[0 : i+1]
				break
			}
		}
		for _, bucket := range list {
			err := q.parseDimensionsAggsData(rs, bucket.Aggregations, append(buckets, bucket))
			if err != nil {
				return err
			}
		}
	} else if rng, ok := aggs.Range("range"); ok {
		for _, bucket := range rng.Buckets {
			err := q.parseDimensionsAggsData(rs, bucket.Aggregations, append(buckets, bucket))
			if err != nil {
				return err
			}
		}
	} else {
		var source map[string]interface{}
		hits, ok := aggs.TopHits("columns")
		if ok && hits != nil && hits.Hits != nil && len(hits.Hits.Hits) > 0 {
			hit := hits.Hits.Hits[0]
			err := json.Unmarshal([]byte(*hit.Source), &source)
			if err != nil {
				return err
			}
		}
		if len(source) <= 0 && len(buckets) <= 0 && q.flag&(queryFlagDimensions|queryFlagAggs) == queryFlagNone {
			return nil
		}
		var values []interface{}
		for _, c := range q.columns {
			v, err := c.getAggValue(source, buckets, aggs)
			if err != nil {
				return err
			}
			values = append(values, v)
		}
		q.ctx.row++
		rs.Rows = append(rs.Rows, values)
	}
	return nil
}

type columnHandler struct {
	field *influxql.Field
	col   *tsql.Column
	ctx   *Context
	fns   map[string]AggHandler
}

func (c *columnHandler) AllColumns() bool {
	if c.field == nil {
		return false
	}
	_, ok := c.field.Expr.(*influxql.Wildcard)
	return ok
}

func (c *columnHandler) getRawValue(source map[string]interface{}) (interface{}, error) {
	return c.getAggFieldExprValue(source, nil, nil, c.field.Expr)
}

func (c *columnHandler) getAggValue(source map[string]interface{}, buckets []interface{}, aggs elastic.Aggregations) (interface{}, error) {
	if c.field == nil {
		if c.col.Flag&tsql.ColumnFlagGroupByInterval == tsql.ColumnFlagGroupByInterval {
			if fn, ok := tsql.BuildInFuntions["time"]; ok {
				v, err := fn(c.ctx, buckets...)
				if err != nil {
					return nil, err
				}
				return v, nil
			}
		} else if c.col.Flag&tsql.ColumnFlagGroupByRange == tsql.ColumnFlagGroupByRange {
			if fn, ok := tsql.BuildInFuntions["range"]; ok {
				v, err := fn(c.ctx, buckets...)
				if err != nil {
					return nil, err
				}
				return v, nil
			}
		}
		return nil, fmt.Errorf("field is nil and invalid column flag %s", c.col.Flag.String())
	}
	return c.getAggFieldExprValue(source, buckets, aggs, c.field.Expr)
}

func (c *columnHandler) getAggFieldExprValue(source map[string]interface{}, buckets []interface{}, aggs elastic.Aggregations, expr influxql.Expr) (interface{}, error) {
	switch expr := expr.(type) {
	case *influxql.Call:
		if fn, ok := tsql.LiteralFuntions[expr.Name]; ok {
			c.col.Flag |= tsql.ColumnFlagFunc | tsql.ColumnFlagLiteral
			var args []interface{}
			for _, arg := range expr.Args {
				arg, err := c.getAggFieldExprValue(source, buckets, aggs, arg)
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
			}
			v, err := fn(c.ctx, args...)
			if err != nil {
				return nil, err
			}
			return v, nil
		} else if fn, ok := tsql.BuildInFuntions[expr.Name]; ok {
			c.col.Flag |= tsql.ColumnFlagFunc
			var args []interface{}
			if expr.Name == "time" || expr.Name == "timestamp" {
				args = buckets
				c.col.Flag |= tsql.ColumnFlagGroupByInterval
			} else if expr.Name == "range" {
				args = buckets
				c.col.Flag |= tsql.ColumnFlagGroupByRange
			} else if expr.Name == "scope" {
				if len(expr.Args) <= 0 {
					return nil, fmt.Errorf("invalid function 'scope' args")
				}
				scope := "terms"
				if len(expr.Args) == 2 {
					sl, _ := expr.Args[1].(*influxql.StringLiteral)
					if len(sl.Val) > 0 {
						scope = sl.Val
					}
				}
				args = append(buckets, scope, expr.Args[0])
			} else {
				for _, arg := range expr.Args {
					v, err := c.getAggFieldExprValue(source, buckets, aggs, arg)
					if err != nil {
						return nil, err
					}
					args = append(args, v)
				}
			}
			v, err := fn(c.ctx, args...)
			if err != nil {
				return nil, err
			}
			return v, nil
		} else if _, ok := AggFunctions[expr.Name]; ok {
			c.col.Flag |= tsql.ColumnFlagFunc | tsql.ColumnFlagAgg
			id := c.ctx.GetFuncID(expr, influxql.AnyField)
			fn, ok := c.fns[id]
			if ok {
				if aggs == nil {
					return nil, nil
				}
				v, err := fn.Handle(aggs)
				if err != nil {
					return nil, err
				}
				return v, nil
			}
		}
	case *influxql.BinaryExpr:
		lv, err := c.getAggFieldExprValue(source, buckets, aggs, expr.LHS)
		if err != nil {
			return nil, err
		}
		rv, err := c.getAggFieldExprValue(source, buckets, aggs, expr.RHS)
		if err != nil {
			return nil, err
		}
		v, err := tsql.OperateValues(lv, toOperator(expr.Op), rv)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *influxql.ParenExpr:
		return c.getAggFieldExprValue(source, buckets, aggs, expr.Expr)
	case *influxql.VarRef:
		key, flag := getKeyNameAndFlag(expr, influxql.AnyField)
		if expr == c.field.Expr {
			c.col.Key = key
		}
		c.col.Flag |= flag
		return getGetValueFromFlatMap(source, key, "."), nil
	default:
		v, ok, err := getLiteralValue(c.ctx, expr)
		if err != nil {
			return nil, err
		}
		if ok {
			c.col.Flag |= tsql.ColumnFlagLiteral
			return v, nil
		}
	}
	return nil, fmt.Errorf("invalid field '%s'", c.field.String())
}

func getColumnName(field *influxql.Field) string {
	if len(field.Alias) > 0 {
		return field.Alias
	}
	return field.String()
}

func getGetValueFromFlatMap(source map[string]interface{}, key string, sep string) interface{} {
	keys := strings.Split(key, sep)
	for i, k := range keys {
		v := source[k]
		if i < len(keys)-1 {
			v, ok := v.(map[string]interface{})
			if !ok {
				return nil
			}
			source = v
			continue
		}
		return v
	}
	return nil
}

func parseSourceColumns(prefix string, source map[string]interface{}, cols map[string]string) {
	if source == nil {
		return
	}
	for key, val := range source {
		k := key
		if len(prefix) > 0 {
			key = prefix + "." + key
		}
		v, ok := val.(map[string]interface{})
		if ok {
			parseSourceColumns(key, v, cols)
			continue
		}
		cols[key] = k
	}
}

// SortResultSet .
func SortResultSet(rs *tsql.ResultSet, sorts influxql.SortFields) {
	type sortitem struct {
		ascending bool
		idx       int
	}
	var idx []*sortitem
	for _, s := range sorts {
		for i, c := range rs.Columns {
			if s.Expr.String() == c.Name {
				idx = append(idx, &sortitem{
					ascending: s.Ascending,
					idx:       i,
				})
			}
		}
	}
	if len(idx) > 0 && len(rs.Rows) > 0 {
		sort.Slice(rs.Rows, func(i, j int) bool {
			a, b := rs.Rows[i], rs.Rows[j]
			for _, s := range idx {
				if a[s.idx] == b[s.idx] {
					continue
				}
				cmp, err := tsql.OperateValues(a[s.idx], tsql.LT, b[s.idx])
				if err != nil {
					continue
				}
				less, ok := cmp.(bool)
				if !ok {
					continue
				}
				if s.ascending {
					return less
				}
				return !less
			}
			return false
		})
	}
}
