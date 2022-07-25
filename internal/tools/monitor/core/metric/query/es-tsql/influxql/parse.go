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
	"strings"

	"github.com/influxdata/influxql"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

// Parser .
type Parser struct {
	debug     bool
	stm       string
	ql        *influxql.Parser
	ctx       *Context
	filter    []*model.Filter
	queryPlan *influxql.Query

	orgName     string
	terminusKey string
}

// New start and end always nanosecond
func New(start, end int64, stmt string, debug bool) tsql.Parser {
	return &Parser{
		stm:   stmt,
		ql:    influxql.NewParser(strings.NewReader(stmt)),
		debug: debug,
		ctx: &Context{
			start:            start,
			end:              end,
			originalTimeUnit: tsql.Nanosecond,
			targetTimeUnit:   tsql.UnsetTimeUnit,
			timeKey:          model.TimestampKey,
			maxTimePoints:    512,
		},
	}
}

func init() {
	tsql.RegisterParser("influxql", New)
}

func (p *Parser) SetOrgName(org string) tsql.Parser {
	p.orgName = org
	return p
}
func (p *Parser) SetTerminusKey(terminusKey string) tsql.Parser {
	p.terminusKey = terminusKey
	return p
}

// SetFilter .
func (p *Parser) SetFilter(filters []*model.Filter) (tsql.Parser, error) {
	p.filter = filters
	return p, nil
}

// SetParams .
func (p *Parser) SetParams(params map[string]interface{}) tsql.Parser {
	if len(params) > 0 {
		p.ql.SetParams(params)
	}
	return p
}

// SetOriginalTimeUnit .
func (p *Parser) SetOriginalTimeUnit(unit tsql.TimeUnit) tsql.Parser {
	p.ctx.originalTimeUnit = unit
	return p
}

// SetTargetTimeUnit .
func (p *Parser) SetTargetTimeUnit(unit tsql.TimeUnit) tsql.Parser {
	p.ctx.targetTimeUnit = unit
	return p
}

// SetTimeKey .
func (p *Parser) SetTimeKey(key string) tsql.Parser {
	p.ctx.timeKey = key
	return p
}

// SetMaxTimePoints .
func (p *Parser) SetMaxTimePoints(points int64) tsql.Parser {
	p.ctx.maxTimePoints = points
	return p
}

func (p *Parser) Build() error {
	if p.ql != nil {
		query, err := p.ql.ParseQuery()
		if err != nil {
			return fmt.Errorf("influxql parse build stmt is error :%v", err)
		}
		p.queryPlan = query

	} else {
		return errors.New("query influxql is not nil")
	}
	return nil
}

func (p *Parser) Metrics() ([]string, error) {
	if p.queryPlan == nil {
		return nil, errors.New("please call Build() first")
	}
	var metrics []string
	for _, stmt := range p.queryPlan.Statements {
		s, ok := stmt.(*influxql.SelectStatement)
		if !ok {
			return nil, tsql.ErrNotSupportNonQueryStatement
		}
		for _, source := range s.Sources {
			measurement, ok := source.(*influxql.Measurement)
			if !ok {
				return nil, errors.New("")
			}
			metrics = append(metrics, measurement.Name)
		}
	}

	return metrics, nil
}

// ParseQuery .
func (p *Parser) ParseQuery(kind string) ([]tsql.Query, error) {
	if p.queryPlan == nil {
		return nil, errors.New("please call Build() first")
	}

	q := p.queryPlan

	var qs []tsql.Query
	for _, stmt := range q.Statements {
		s, ok := stmt.(*influxql.SelectStatement)
		if !ok {
			return nil, tsql.ErrNotSupportNonQueryStatement
		}
		var q tsql.Query
		var err error
		if kind == model.ClickhouseKind {
			q, err = p.ParseClickhouse(s)
		} else {
			q, err = p.ParseElasticsearch(s)
		}
		if err != nil {
			return nil, err
		}
		qs = append(qs, q)

	}
	return qs, nil
}

func getKeyName(ref *influxql.VarRef, deftyp influxql.DataType) string {
	name, _ := getKeyNameAndFlag(ref, deftyp)
	return name
}

const nameKey = "_" + model.NameKey

func getKeyNameAndFlag(ref *influxql.VarRef, deftyp influxql.DataType) (string, model.ColumnFlag) {
	if ref.Type == influxql.Unknown {
		if ref.Val == model.TimestampKey || ref.Val == model.TimeKey {
			return model.TimestampKey, model.ColumnFlagTimestamp
		} else if ref.Val == model.NameKey || ref.Val == nameKey {
			return model.NameKey, model.ColumnFlagName
		}
		if deftyp == influxql.Tag {
			return model.TagsKey + ref.Val, model.ColumnFlagTag
		}
	} else if ref.Type == influxql.Tag {
		return model.TagsKey + ref.Val, model.ColumnFlagTag
	}
	return model.FieldsKey + ref.Val, model.ColumnFlagField
}

// getLiteralValue try to get constants.
func getLiteralValue(ctx *Context, expr influxql.Expr) (interface{}, bool, error) {
	switch val := expr.(type) {
	case *influxql.IntegerLiteral:
		return val.Val, true, nil
	case *influxql.NumberLiteral:
		return val.Val, true, nil
	case *influxql.UnsignedLiteral:
		return val.Val, true, nil
	case *influxql.BooleanLiteral:
		return val.Val, true, nil
	case *influxql.StringLiteral:
		return val.Val, true, nil
	case *influxql.DurationLiteral:
		return int64(val.Val), true, nil
	case *influxql.TimeLiteral:
		return val.Val.UnixNano(), true, nil
	case *influxql.RegexLiteral:
		return val.Val, true, nil
	case *influxql.NilLiteral:
		return nil, true, nil
	case *influxql.ListLiteral:
		return val.Vals, true, nil
	case *influxql.Call:
		return getLiteralFuncValue(ctx, val)
	case *influxql.BinaryExpr:
		lv, ok, err := getLiteralValue(ctx, val.LHS)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		rv, ok, err := getLiteralValue(ctx, val.RHS)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		v, err := tsql.OperateValues(lv, toOperator(val.Op), rv)
		if err != nil {
			return nil, false, err
		}
		return v, true, nil
	case *influxql.ParenExpr:
		return getLiteralValue(ctx, val.Expr)
	}
	return nil, false, nil
}

func getLiteralFuncValue(ctx *Context, call *influxql.Call) (interface{}, bool, error) {
	if fn, ok := tsql.LiteralFunctions[call.Name]; ok {
		var args []interface{}
		for _, arg := range call.Args {
			arg, ok, err := getLiteralValue(ctx, arg)
			if err != nil {
				return nil, false, err
			}
			if !ok {
				return nil, false, fmt.Errorf("invalid args in literal function '%s'", call.Name)
			}
			args = append(args, arg)
		}
		v, err := fn(ctx, args...)
		if err != nil {
			return nil, false, err
		}
		return v, true, nil
	} else if fn, ok := tsql.BuildInFunctions[call.Name]; ok {
		var args []interface{}
		for _, arg := range call.Args {
			arg, ok, err := getLiteralValue(ctx, arg)
			if err != nil {
				return nil, false, err
			}
			if ok {
				return nil, false, nil
			}
			args = append(args, arg)
		}
		v, err := fn(ctx, args...)
		if err != nil {
			return nil, false, err
		}
		return v, true, nil
	}
	return nil, false, nil
}
