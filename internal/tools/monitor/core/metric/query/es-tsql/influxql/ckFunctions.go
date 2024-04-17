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
	"math"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/influxdata/influxql"

	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

type SQLAggHandler interface {
	Aggregations(p *Parser, aggs map[string]exp.Expression, flags ...FuncFlag) error
	Handle(row map[string]interface{}) (interface{}, error)
}

type SQlAggFuncDefine struct {
	Flag FuncFlag
	New  func(ctx *Context, id string, call *influxql.Call) (SQLAggHandler, error)
}

var CkAggFunctions = map[string]*SQlAggFuncDefine{
	"max": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"max",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.MAX(lit).As(id), nil
				}

				f := p.ckColumnByOnlyExistingColumn(field)
				return goqu.MAX(goqu.L(f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"min": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"min",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.MIN(lit).As(id), nil
				}
				f := p.ckColumnByOnlyExistingColumn(field)
				return goqu.MIN(goqu.L(f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"avg": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"avg",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.AVG(lit).As(id), nil
				}
				f := p.ckColumnByOnlyExistingColumn(field)
				return goqu.AVG(goqu.L(f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				if v == nil {
					return 0, true
				}
				float64V, ok := v.(float64)
				if ok && math.IsNaN(float64V) {
					return 0, true
				}
				return v, true
			},
		),
	},
	"sum": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"sum",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.SUM(lit).As(id), nil
				}
				f, _ := p.ckGetKey(field, influxql.AnyField)
				return goqu.SUM(goqu.L(f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"count": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"count",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.COUNT(lit).As(id), nil
				}
				f := p.ckColumnByOnlyExistingColumn(field)
				return goqu.COUNT(goqu.L(f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"distinct": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"distinct",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.DISTINCT(lit).As(id), nil
				}
				f := p.ckColumnByOnlyExistingColumn(field)
				return goqu.L(fmt.Sprintf("uniqCombined(%s)", f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"diff": {
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"diff",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				var f string
				if lit != nil {
					if litExpression, ok := lit.(exp.LiteralExpression); ok {
						f = litExpression.Literal()
					} else {
						return goqu.MIN(lit).As(id), nil
					}
				} else {
					f = p.ckColumnByOnlyExistingColumn(field)
				}
				return goqu.L(fmt.Sprintf("Min(%s)", f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				if v == nil {
					v = float64(0)
				}
				if next, ok := ctx.attributesCache["next"]; ok {
					currentV, ok := v.(float64)
					if !ok {
						return nil, false
					}
					if nextV, ok := next.(map[string]interface{}); ok {
						if nextV[id] != nil {
							nextValue := nextV[id].(float64)
							return nextValue - currentV, true
						}
					}
				}
				return 0, true
			},
		),
	},
	"diffps": {
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"diffps",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				var f string
				if lit != nil {
					if litExpression, ok := lit.(exp.LiteralExpression); ok {
						f = litExpression.Literal()
					} else {
						return goqu.MIN(lit).As(id), nil
					}
				} else {
					f = p.ckColumnByOnlyExistingColumn(field)
				}
				return goqu.L(fmt.Sprintf("Min(%s)", f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				if v == nil {
					v = float64(0)
				}
				if next, ok := ctx.attributesCache["next"]; ok {
					currentV, ok := v.(float64)
					if !ok {
						return nil, false
					}
					if nextV, ok := next.(map[string]interface{}); ok {
						if nextV[id] != nil {
							if ctx.targetTimeUnit == tsql.UnsetTimeUnit {
								ctx.targetTimeUnit = tsql.Nanosecond
							}
							seconds := float64(ctx.interval*int64(ctx.targetTimeUnit)) / float64(tsql.Second)
							nextValue := nextV[id].(float64)
							// diffps return value should not be negative
							// erda issue: https://erda.cloud/erda/dop/projects/387/issues/all?id=581160&iterationID=12783&tab=BUG&type=BUG
							if nextValue < currentV {
								return 0, true
							}
							return (nextValue - currentV) / seconds, true
						}
					}
				}
				return 0, true
			},
		),
	},
	"rateps": {
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"rateps",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.SUM(field).As(id), nil
				}
				f, _ := p.ckGetKey(field, influxql.AnyField)
				return goqu.SUM(goqu.L(f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				if v == nil {
					v = float64(0)
				}
				if currentV, ok := v.(float64); ok {
					if ctx.targetTimeUnit == tsql.UnsetTimeUnit {
						ctx.targetTimeUnit = tsql.Nanosecond
					}
					seconds := float64(ctx.interval*int64(ctx.targetTimeUnit)) / float64(tsql.Second)
					return currentV / seconds, true
				}
				return nil, false
			},
		),
	},
	"value": {
		// last value
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"value",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.MAX(lit).As(id), nil
				}
				f, _ := p.ckGetKey(field, influxql.AnyField)
				return goqu.MAX(goqu.L(f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"first": {
		// last value
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"value",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				var f string
				if lit != nil {
					if litExpression, ok := lit.(exp.LiteralExpression); ok {
						f = litExpression.Literal()
					} else {
						return goqu.MIN(lit).As(id), nil
					}
				} else {
					f = p.ckColumnByOnlyExistingColumn(field)
				}
				return goqu.L(fmt.Sprintf("argMin(%s,%s)", f, ctx.TimeKey())).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"last": {
		// last value
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"last",
			func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				var f string
				if lit != nil {
					if litExpression, ok := lit.(exp.LiteralExpression); ok {
						f = litExpression.Literal()
					} else {
						return goqu.MAX(lit).As(id), nil
					}
				} else {
					f = p.ckColumnByOnlyExistingColumn(field)
				}
				return goqu.L(fmt.Sprintf("anyLast(%s)", f)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
}

type ckUnaryFunction struct {
	name   string
	id     string
	field  string
	call   *influxql.Call
	ctx    *Context
	agg    func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error)
	getter func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool)
}

func newCkUnaryFunction(
	name string,
	agg func(ctx *Context, p *Parser, id string, field *influxql.VarRef, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error),
	getter func(ctx *Context, id, field string, call *influxql.Call, aggs interface{}) (interface{}, bool),
) func(ctx *Context, id string, call *influxql.Call) (SQLAggHandler, error) {
	return func(ctx *Context, id string, call *influxql.Call) (SQLAggHandler, error) {
		return &ckUnaryFunction{
			id:     id,
			name:   name,
			ctx:    ctx,
			call:   call,
			agg:    agg,
			getter: getter,
		}, nil
	}
}

func (c ckUnaryFunction) Aggregations(p *Parser, aggs map[string]exp.Expression, flags ...FuncFlag) error {
	call := c.call
	if err := mustCallArgsNum(call, 1); err != nil {
		return err
	}
	arg := call.Args[0]
	c.id = c.ctx.GetFuncID(call, influxql.AnyField)
	if _, ok := aggs[c.id]; ok {
		return nil
	}
	if ref, ok := arg.(*influxql.VarRef); ok {
		c.field = ref.Val
		id := c.id
		a, err := c.agg(c.ctx, p, id, ref, nil, flags...)
		if err != nil {
			return err
		}
		aggs[c.id] = a
	} else {
		script, err := p.getScriptExpressionOnCk(c.ctx, arg, influxql.AnyField, nil)
		if err != nil {
			return err
		}
		a, err := c.agg(c.ctx, p, c.id, nil, goqu.L(script), flags...)
		if err != nil {
			return err
		}
		aggs[c.id] = a
	}
	return nil
}

func (c ckUnaryFunction) Handle(row map[string]interface{}) (interface{}, error) {
	val, ok := c.getter(c.ctx, c.id, c.field, c.call, row[c.id])
	if !ok {
		return nil, fmt.Errorf("invalid %s Aggregation %s", c.name, c.id)
	}
	return val, nil
}
