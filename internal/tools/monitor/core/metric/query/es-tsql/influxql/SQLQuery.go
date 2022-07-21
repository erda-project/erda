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
	"fmt"

	"github.com/influxdata/influxql"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

type SQLColumnHandler struct {
	field *influxql.Field
	col   *model.Column
	ctx   *Context
	fns   map[string]SQLAggHandler
}

func (c *SQLColumnHandler) AllColumns() bool {
	if c.field == nil {
		return false
	}
	_, ok := c.field.Expr.(*influxql.Wildcard)
	return ok
}

func (c *SQLColumnHandler) getValue(currentRow map[string]interface{}) (interface{}, error) {
	if c.field == nil {
		return nil, nil
	}
	return c.getAggFieldExprValue(currentRow, c.field.Expr)
}

func (c *SQLColumnHandler) getAggFieldExprValue(row map[string]interface{}, exprr influxql.Expr) (interface{}, error) {
	switch expr := exprr.(type) {
	case *influxql.Call:
		if fn, ok := tsql.LiteralFunctions[expr.Name]; ok {
			c.col.Flag |= model.ColumnFlagFunc | model.ColumnFlagLiteral
			var args []interface{}
			for _, arg := range expr.Args {
				arg, err := c.getAggFieldExprValue(row, arg)
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
		} else if fn, ok := tsql.CkBuildInFunctions[expr.Name]; ok {
			if expr.Name == "time" || expr.Name == "timestamp" {
				v := getGetValueFromFlatMap(row, "bucket_timestamp", ".")
				return fn(c.ctx, v)
			}
			c.col.Flag |= model.ColumnFlagFunc
			var args []interface{}
			for _, arg := range expr.Args {
				v, err := c.getAggFieldExprValue(row, arg)
				if err != nil {
					return nil, err
				}
				args = append(args, v)
			}

			v, err := fn(c.ctx, args...)
			if err != nil {
				return nil, err
			}
			return v, nil
		} else if _, ok := CkAggFunctions[expr.Name]; ok {
			c.col.Flag |= model.ColumnFlagFunc | model.ColumnFlagAgg
			id := c.ctx.GetFuncID(expr, influxql.AnyField)
			fn, ok := c.fns[id]
			if ok {
				if row == nil {
					return nil, nil
				}
				v, err := fn.Handle(row)
				if err != nil {
					return nil, err
				}
				return v, nil
			}
		}
	case *influxql.BinaryExpr:
		lv, err := c.getAggFieldExprValue(row, expr.LHS)
		if err != nil {
			return nil, err
		}
		rv, err := c.getAggFieldExprValue(row, expr.RHS)
		if err != nil {
			return nil, err
		}
		v, err := tsql.OperateValues(lv, toOperator(expr.Op), rv)
		if err != nil {
			return nil, err
		}
		return v, nil
	case *influxql.ParenExpr:
		return c.getAggFieldExprValue(row, expr.Expr)
	case *influxql.VarRef:
		key := expr.Val

		if expr == c.field.Expr {
			v := c.field.Expr.(*influxql.VarRef)
			key = v.Val
		}
		return getGetValueFromFlatMap(row, key, "."), nil
	default:
		v, ok, err := getLiteralValue(c.ctx, expr)
		if err != nil {
			return nil, err
		}
		if ok {
			c.col.Flag |= model.ColumnFlagLiteral
			return v, nil
		}
	}
	return nil, fmt.Errorf("invalid field '%s'", c.field.String())
}
