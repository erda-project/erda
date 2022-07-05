package esinfluxql

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/influxdata/influxql"

	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

type SQLAggHandler interface {
	Aggregations(aggs map[string]exp.Expression, flags ...FuncFlag) error
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
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.MAX(lit).As(id), nil
				}
				return goqu.MAX(goqu.L(field)).As(id), nil
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
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.MIN(lit).As(id), nil
				}
				return goqu.MIN(goqu.L(field)).As(id), nil
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
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.AVG(lit).As(id), nil
				}
				return goqu.AVG(goqu.L(field)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return true, true
			},
		),
	},
	"sum": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"sum",
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.SUM(lit).As(id), nil
				}
				return goqu.SUM(goqu.L(field)).As(id), nil
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
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.COUNT(lit).As(id), nil
				}
				return goqu.COUNT(goqu.L(field)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				return v, true
			},
		),
	},
	"distinct": {
		Flag: FuncFlagSelect | FuncFlagOrderBy,
		New: newCkUnaryFunction(
			"",
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.DISTINCT(lit).As(id), nil
				}
				return goqu.DISTINCT(goqu.L(field)).As(id), nil
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
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.MIN(lit).As(id), nil
				}
				return goqu.MIN(goqu.L(field)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				if next, ok := ctx.attributesCache["next"]; ok {
					currentV, ok := v.(float64)
					if !ok {
						return nil, false
					}
					if nextV, ok := next.(map[string]interface{}); ok {
						if nextV[field] != nil {
							return nextV[field].(float64) - currentV, true
						}
					}
				}
				return nil, false
			},
		),
	},
	"diffps": {
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"diffps",
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.MIN(field).As(id), nil
				}
				return goqu.MIN(goqu.L(field)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
				//speed
				if next, ok := ctx.attributesCache["next"]; ok {
					currentV, ok := v.(float64)
					if !ok {
						return nil, false
					}
					if nextV, ok := next.(map[string]interface{}); ok {
						if nextV[field] != nil {
							if ctx.targetTimeUnit == tsql.UnsetTimeUnit {
								ctx.targetTimeUnit = tsql.Nanosecond
							}
							seconds := float64(ctx.interval*int64(ctx.targetTimeUnit)) / float64(tsql.Second)
							return (nextV[field].(float64) - currentV) / seconds, true
						}
					}
				}
				return nil, false
			},
		),
	},
	"rateps": {
		Flag: FuncFlagSelect,
		New: newCkUnaryFunction(
			"rateps",
			func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error) {
				if lit != nil {
					return goqu.SUM(field).As(id), nil
				}
				return goqu.SUM(goqu.L(field)).As(id), nil
			},
			func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool) {
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
}

type ckUnaryFunction struct {
	name   string
	id     string
	field  string
	call   *influxql.Call
	ctx    *Context
	agg    func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error)
	getter func(ctx *Context, id, field string, call *influxql.Call, v interface{}) (interface{}, bool)
}

func newCkUnaryFunction(
	name string,
	agg func(ctx *Context, id, field string, lit exp.Expression, flags ...FuncFlag) (exp.Expression, error),
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

func (c ckUnaryFunction) Aggregations(aggs map[string]exp.Expression, flags ...FuncFlag) error {
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
		c.field = ckGetKeyName(ref, influxql.AnyField)
		id := c.id
		a, err := c.agg(c.ctx, id, c.field, nil, flags...)
		if err != nil {
			return err
		}
		aggs[c.id] = a
	} else {
		script, err := getScriptExpressionOnCk(c.ctx, arg, influxql.AnyField, nil)
		if err != nil {
			return err
		}
		a, err := c.agg(c.ctx, c.id, "", goqu.L(script), flags...)
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
