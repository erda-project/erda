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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/influxdata/influxql"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

func (p *Parser) ParseClickhouse(s *influxql.SelectStatement) (tsql.Query, error) {

	sources, err := p.from(s.Sources)

	if err != nil {
		return nil, errors.Wrap(err, "select stmt parse to from is error")
	}

	expr := goqu.From("metric") // metric is fake, in execution layer it's real table

	expr = p.appendTimeKeyByExpr(expr)

	// add parser filter to expr
	expr, err = p.filterOnExpr(expr)
	if err != nil {
		return nil, errors.Wrap(err, "select stmt parse to filter is error")
	}

	expr, err = p.conditionOnExpr(expr, s)
	if err != nil {
		return nil, errors.Wrap(err, "select stmt parse to condition is error")
	}

	// select
	expr, handlers, columns, err := p.parseQueryOnExpr(s.Fields, expr)
	if err != nil {
		return nil, errors.Wrap(err, "select stmt parse to select is error")
	}

	expr, tailLiters, err := p.ParseGroupByOnExpr(s.Dimensions, expr, &handlers, columns)
	if err != nil {
		return nil, errors.Wrap(err, "select stmt parse to group by is error")
	}

	expr, err = p.ParseOrderByOnExpr(s.SortFields, expr, columns)
	if err != nil {
		return nil, errors.Wrap(err, "select stmt parse to order by is error")
	}

	if len(tailLiters) > 0 {
		expr = expr.OrderAppend(goqu.I("%s").Asc()) //Placeholder!!!
	}
	expr, err = p.ParseOffsetAndLimitOnExpr(s, expr)
	if err != nil {
		return nil, errors.Wrap(err, "select stmt parse to offset and limit is error")
	}
	return QueryClickhouse{
		sources:     sources,
		subLiters:   tailLiters,
		column:      handlers,
		start:       p.ctx.start,
		end:         p.ctx.end,
		kind:        model.ClickhouseKind,
		ctx:         p.ctx,
		expr:        expr,
		debug:       p.debug,
		orgName:     p.orgName,
		terminusKey: p.terminusKey,
	}, nil
}

func (p *Parser) ParseOrderByOnExpr(s influxql.SortFields, expr *goqu.SelectDataset, columns map[string]string) (*goqu.SelectDataset, error) {
	if len(s) <= 0 {
		return expr, nil
	}
	sortFields := make(map[string]bool)

	if len(s) <= 0 {
		sortFields[p.ctx.timeKey] = false
	} else {
		for _, field := range s {
			if field.Expr == nil {
				sortFields[p.ctx.timeKey] = field.Ascending
				continue
			}

			if v, ok := field.Expr.(*influxql.VarRef); ok {
				c, _ := p.ckGetKeyName(v, influxql.AnyField)
				sortFields[c] = field.Ascending
			} else {
				script, err := getAggsOrderScript(p.ctx, field.Expr)
				if err != nil {
					return nil, err
				}
				sortFields[script] = field.Ascending
			}
		}
	}

	for key, isAsc := range sortFields {
		column := key
		if newName, ok := columns[key]; ok {
			column = newName
		}
		// simple column
		if !isAsc {
			expr = expr.OrderAppend(goqu.C(column).Desc())
		} else {
			expr = expr.OrderAppend(goqu.C(column).Asc())
		}
		continue
	}

	return expr, nil
}

func getAggsOrderScript(ctx *Context, expr influxql.Expr) (string, error) {
	switch expr := expr.(type) {
	case *influxql.Call:
		if define, ok := CkAggFunctions[expr.Name]; ok {
			if define.Flag&FuncFlagOrderBy == 0 {
				return "", fmt.Errorf("not support function '%s' in order by", expr.Name)
			}
			id := ctx.GetFuncID(expr, influxql.AnyField)
			return id, nil
		}
	case *influxql.ParenExpr:
		return getAggsOrderScript(ctx, expr.Expr)
	}
	return "", fmt.Errorf("invalid order by expression")
}

func (p *Parser) ParseOffsetAndLimitOnExpr(s *influxql.SelectStatement, expr *goqu.SelectDataset) (*goqu.SelectDataset, error) {
	if s.Limit > 0 {
		expr = expr.Limit(uint(s.Limit))
	}
	if s.Offset <= 0 {
		s.Offset = 0
	}
	expr = expr.Offset(uint(s.Offset))
	return expr, nil
}

func (p *Parser) ParseGroupByOnExpr(dimensions influxql.Dimensions, expr *goqu.SelectDataset, handlers *[]*SQLColumnHandler, columns map[string]string) (*goqu.SelectDataset, map[string]string, error) {
	if len(dimensions) <= 0 {
		return expr, nil, nil
	}
	expr, liters, tailLiters, err := p.parseQueryDimensionsByExpr(expr, dimensions, handlers)
	if err != nil {
		return nil, nil, err
	}

	groupExpress := make(map[string]goqu.Expression)
	if len(liters) > 0 {
		for script, columnName := range columns {
			groupExpress[script] = goqu.C(columnName)
		}
		for _, liter := range liters {
			if _, ok := groupExpress[liter]; !ok {
				groupExpress[liter] = goqu.L(liter)
			}
		}
	}
	for _, express := range groupExpress {
		expr = expr.GroupByAppend(express)
	}
	return expr, tailLiters, nil
}
func (p *Parser) parseQueryOnExpr(fields influxql.Fields, expr *goqu.SelectDataset) (*goqu.SelectDataset, []*SQLColumnHandler, map[string]string, error) {
	columns := make(map[string]string) // k:stmt, v: column
	aggs, handlers, err := p.parseQueryFieldsByExpr(fields, columns)
	if err != nil {
		return expr, nil, nil, err
	}

	expr = expr.Select(nil)
	for _, column := range aggs {
		expr = expr.SelectAppend(column)
	}

	for column, asName := range columns {
		if len(asName) <= 0 {
			expr = expr.SelectAppend(goqu.L(column))
			continue
		}
		expr = expr.SelectAppend(goqu.L(column).As(asName))
	}
	return expr, handlers, columns, nil
}
func (p *Parser) parseQueryDimensionsByExpr(exprSelect *goqu.SelectDataset, dimensions influxql.Dimensions, handler *[]*SQLColumnHandler) (*goqu.SelectDataset, []string, map[string]string, error) {
	var exprList []string

	tailExpr := make(map[string]string)
	for _, dim := range dimensions {
		switch expr := dim.Expr.(type) {
		case *influxql.Call:
			if expr.Name == "time" {
				var interval int64
				if len(tailExpr) > 0 {
					return nil, nil, nil, fmt.Errorf("with fill statement should be one function")
				}
				if len(expr.Args) == 1 {
					arg := expr.Args[0]
					d, ok := arg.(*influxql.DurationLiteral)
					if !ok || d.Val < time.Second {
						return exprSelect, nil, nil, fmt.Errorf("invalid arg '%s' in function '%s'", arg.String(), expr.Name)
					}
					fmt.Println(d.Val.String())
					interval = int64(d.Val)
				}
				start, end := p.ctx.Range(true)
				interval = adjustInterval(start, end, interval, p.ctx.maxTimePoints)
				p.ctx.interval = interval

				if p.ctx.OriginalTimeUnit() != tsql.UnsetTimeUnit {
					interval /= int64(p.ctx.OriginalTimeUnit())
				}
				if p.ctx.TargetTimeUnit() != tsql.UnsetTimeUnit {
					p.ctx.interval /= int64(p.ctx.TargetTimeUnit())
				}

				if len(p.ctx.TimeKey()) <= 0 {
					p.ctx.timeKey = model.TimestampKey
				}

				timeBucketColumn := fmt.Sprintf("bucket_%s", p.ctx.TimeKey())
				intervalSeconds := interval / int64(tsql.Second)

				bucketStartTime := (start / interval) * interval
				bucketEndTime := (end / interval) * interval

				exprSelect = exprSelect.SelectAppend(goqu.L(fmt.Sprintf("toDateTime64(toStartOfInterval(timestamp, toIntervalSecond(%v)),9)", intervalSeconds)).As(timeBucketColumn))

				// todo goqu order statement should be literal + asc, but with fill is order by `column` [asc/desc] with fill from %left to %right step %interval
				// buck_timestamp with fill from  fromUnixTimestamp64Nano(cast(1657584000000000000, 'Int64')) to fromUnixTimestamp64Nano(cast(1657756800000000000, 'Int64')) step toDateTime64(86400,9)
				tailExpr[timeBucketColumn] = fmt.Sprintf("%s with fill from fromUnixTimestamp64Nano(cast(%v, 'Int64')) to fromUnixTimestamp64Nano(cast(%v, 'Int64')) step  toDateTime64(%v,9)", timeBucketColumn, bucketStartTime, bucketEndTime, intervalSeconds)

				exprList = append(exprList, timeBucketColumn)

				// append to head
				var newHandler []*SQLColumnHandler

				newHandler = append(newHandler, &SQLColumnHandler{
					field: &influxql.Field{
						Expr:  expr,
						Alias: "time",
					},
					ctx: p.ctx,
				})
				newHandler = append(newHandler, *handler...)
				*handler = newHandler
				continue
			} else if expr.Name == "range" {
				continue
			}

			script, err := p.getScriptExpressionOnCk(p.ctx, dim.Expr, influxql.AnyField, nil)
			if err != nil {
				return exprSelect, nil, nil, nil
			}
			exprList = append(exprList, script)
		}
		script := p.getExprStringAndFlagByExpr(dim.Expr, influxql.AnyField)
		exprList = append(exprList, script)
	}
	return exprSelect, exprList, tailExpr, nil
}

func (p *Parser) getExprArgsOnRange(expr *influxql.Call) ([][]int64, error) {
	var arr [][]int64
	var from interface{}
	for i, item := range expr.Args[1:] {
		val, ok, err := getLiteralValue(p.ctx, item)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("args[%d] is literal in 'range' function", i)
		}
		if i%2 == 0 {
			from = val
			continue
		}
		var rang []int64
		rang = append(rang, from.(int64))
		rang = append(rang, val.(int64))
		arr = append(arr, rang)
		from = nil
	}
	if from != nil {
		var rang []int64
		rang = append(rang, from.(int64))
		arr = append(arr, rang)
		from = nil
	}
	return arr, nil
}

func (p *Parser) getExprStringAndFlagByExpr(expr influxql.Expr, deftyp influxql.DataType) (key string) {
	if expr == nil {
		return ""
	}
	switch expr := expr.(type) {
	case *influxql.BinaryExpr:
		left := p.getExprStringAndFlagByExpr(expr.LHS, deftyp)
		right := p.getExprStringAndFlagByExpr(expr.RHS, deftyp)
		return left + expr.Op.String() + right
	case *influxql.Call:
		var args []string
		for _, arg := range expr.Args {
			k := p.getExprStringAndFlagByExpr(arg, deftyp)
			args = append(args, k)
		}
		return expr.Name + "(" + strings.Join(args, ",") + ")"
	case *influxql.ParenExpr:
		key = p.getExprStringAndFlagByExpr(expr.Expr, deftyp)
		return key
	case *influxql.IntegerLiteral:
		return strconv.FormatInt(expr.Val, 10)
	case *influxql.NumberLiteral:
		return strconv.FormatFloat(expr.Val, 'f', -1, 64)
	case *influxql.BooleanLiteral:
		return strconv.FormatBool(expr.Val)
	case *influxql.UnsignedLiteral:
		return strconv.FormatUint(expr.Val, 10)
	case *influxql.StringLiteral, *influxql.NilLiteral, *influxql.TimeLiteral, *influxql.DurationLiteral, *influxql.RegexLiteral, *influxql.ListLiteral:
		return expr.String()
	case *influxql.VarRef:
		key, _, _ = p.ckGetKeyNameAndFlag(expr, deftyp)
		return key
	}
	return expr.String()
}

func (p *Parser) parseQueryFieldsByExpr(fields influxql.Fields, columns map[string]string) (map[string]exp.Expression, []*SQLColumnHandler, error) {
	aggs := make(map[string]exp.Expression)
	var handlers []*SQLColumnHandler
	for _, field := range fields {
		h, err := p.parseFieldByExpr(field, aggs, columns)
		if err != nil {
			return nil, nil, err
		}
		handlers = append(handlers, h)
	}
	return aggs, handlers, nil

}

func (p *Parser) parseFieldByExpr(field *influxql.Field, aggs map[string]exp.Expression, cols map[string]string) (*SQLColumnHandler, error) {
	ch := &SQLColumnHandler{
		field: field,
		ctx:   p.ctx,
	}
	ch.fns = make(map[string]SQLAggHandler)
	err := p.parseFiledAggByExpr(field.Expr, aggs, ch.fns)
	if err != nil {
		return nil, err
	}
	err = p.parseFiledRefByExpr(field.Expr, cols)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (p *Parser) parseFiledAggByExpr(expr influxql.Expr, aggs map[string]exp.Expression, fns map[string]SQLAggHandler) error {
	switch expr := expr.(type) {
	case *influxql.Call:
		if expr.Name == "scope" {
			return p.parseScopeAgg(expr)
		}
		if define, ok := CkAggFunctions[expr.Name]; ok {
			if define.Flag&FuncFlagSelect == 0 {
				return fmt.Errorf("not support function '%s' in select", expr.Name)
			}
			id := p.ctx.GetFuncID(expr, influxql.AnyField)
			if _, ok := fns[id]; !ok {
				fn, err := define.New(p.ctx, id, expr)
				if err != nil {
					return err
				}
				err = fn.Aggregations(p, aggs, FuncFlagSelect)
				if err != nil {
					return err
				}
				fns[id] = fn
			}
		} else if _, ok := tsql.CkBuildInFunctions[expr.Name]; ok {
			for _, arg := range expr.Args {
				if err := p.parseFiledAggByExpr(arg, aggs, fns); err != nil {
					return err
				}
			}
		}
	case *influxql.BinaryExpr:
		if err := p.parseFiledAggByExpr(expr.LHS, aggs, fns); err != nil {
			return err
		}
		if err := p.parseFiledAggByExpr(expr.RHS, aggs, fns); err != nil {
			return err
		}
	case *influxql.ParenExpr:
		return p.parseFiledAggByExpr(expr.Expr, aggs, fns)
	}
	return nil
}

func (p *Parser) parseFiledRefByExpr(expr influxql.Expr, cols map[string]string) error {
	switch expr := expr.(type) {
	case *influxql.Call:
		if expr.Name == "scope" {
			return nil
		}
		_, ok := CkAggFunctions[expr.Name]
		if !ok {
			for _, arg := range expr.Args {
				if err := p.parseFiledRefByExpr(arg, cols); err != nil {
					return err
				}
			}
		}
	case *influxql.BinaryExpr:
		if err := p.parseFiledRefByExpr(expr.LHS, cols); err != nil {
			return err
		}
		if err := p.parseFiledRefByExpr(expr.RHS, cols); err != nil {
			return err
		}
	case *influxql.ParenExpr:
		return p.parseFiledRefByExpr(expr.Expr, cols)
	case *influxql.VarRef:
		c, _ := p.ckGetKeyName(expr, influxql.AnyField)
		cols[c] = expr.Val
	case *influxql.Wildcard:
		cols["*"] = ""
	}
	return nil
}

func (p *Parser) appendTimeKeyByExpr(expr *goqu.SelectDataset) *goqu.SelectDataset {
	start, end := p.ctx.Range(true)
	expr = expr.Where(
		goqu.C(p.ctx.timeKey).Gte(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", start)),
		goqu.C(p.ctx.timeKey).Lt(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", end)),
	)
	return expr
}

func (p *Parser) from(sources influxql.Sources) ([]*model.Source, error) {
	var list []*model.Source
	for _, source := range sources {
		switch s := source.(type) {
		case *influxql.Measurement:
			if s.Regex != nil {
				return nil, errors.New("no support regex")
			}
			list = append(list, &model.Source{
				Database: s.Database,
				Name:     s.Name,
			})
		default:
			return nil, fmt.Errorf("no support from type %s", source.String())
		}
	}
	return list, nil
}

func (p *Parser) filterToExpr(filters []*model.Filter, expr *goqu.SelectDataset) (*goqu.SelectDataset, error) {
	or := goqu.Or()
	expressionList := goqu.And()
	for _, item := range filters {
		key := item.Key
		keyArr := strings.Split(key, ".")
		if len(keyArr) > 1 && keyArr[0] == "tags" {
			key, _ = p.ckGetKeyName(&influxql.VarRef{
				Val:  keyArr[1],
				Type: influxql.Tag,
			}, influxql.Tag)
		}

		switch item.Operator {
		case "eq", "=", "":
			expressionList = expressionList.Append(goqu.L(key).Eq(item.Value))
		case "neq", "!=":
			expressionList = expressionList.Append(goqu.L(key).Neq(item.Value))
		case "gt", ">":
			expressionList = expressionList.Append(goqu.L(key).Gt(item.Value))
		case "gte", ">=":
			expressionList = expressionList.Append(goqu.L(key).Gte(item.Value))
		case "lt", "<":
			expressionList = expressionList.Append(goqu.L(key).Lt(item.Value))
		case "lte", "<=":
			expressionList = expressionList.Append(goqu.L(key).Lte(item.Value))
		case "in":
			if values, ok := item.Value.([]interface{}); ok {
				expressionList = expressionList.Append(goqu.L(key).In(values))
			}
		case "match":
			expressionList = expressionList.Append(goqu.L(key).Like(fmt.Sprintf("%%%v%%", item.Value)))
		case "nmatch":
			expressionList = expressionList.Append(goqu.L(key).NotLike(fmt.Sprintf("%%%v%%", item.Value)))

		case "or_eq":
			orExpr := goqu.L(key).Eq(item.Value)

			if or.IsEmpty() {
				or = goqu.Or(orExpr)
			} else {
				or = or.Append(orExpr)
			}

		case "or_in":
			if values, ok := item.Value.([]interface{}); ok {
				orExpr := goqu.L(key).In(values)
				if or.IsEmpty() {
					or = goqu.Or(orExpr)
				} else {
					or = or.Append(orExpr)
				}
			}
		default:
			return nil, fmt.Errorf("not support filter operator %s", item.Operator)
		}
	}

	if !or.IsEmpty() {
		expr = expr.Where(goqu.Or(
			expressionList,
			or,
		))
		return expr, nil
	}
	if !expressionList.IsEmpty() {
		expr = expr.Where(expressionList)
	}

	return expr, nil
}

func (p *Parser) filterOnExpr(expr *goqu.SelectDataset) (*goqu.SelectDataset, error) {
	var err error
	if len(p.filter) > 0 {
		expr, err = p.filterToExpr(p.filter, expr)
		if err != nil {
			return nil, fmt.Errorf("parse filter to expr error: %v", err)
		}
	}
	return expr, nil
}

func (p *Parser) conditionOnExpr(expr *goqu.SelectDataset, s *influxql.SelectStatement) (*goqu.SelectDataset, error) {
	if s.Condition != nil {
		exprList := goqu.And()
		var err error
		exprList, err = p.parseConditionOnExpr(s.Condition, exprList)
		if err != nil {
			return nil, err
		}
		expr = expr.Where(exprList)
	}
	return expr, nil
}

func (p *Parser) parseConditionOnExpr(cond influxql.Expr, exprList exp.ExpressionList) (exp.ExpressionList, error) {
	var err error
	switch expr := cond.(type) {
	case *influxql.BinaryExpr:
		if expr.Op == influxql.AND || expr.Op == influxql.OR {
			left := goqu.And()
			left, err = p.parseConditionOnExpr(expr.LHS, left)
			if err != nil {
				return exprList, err
			}
			right := goqu.And()
			right, err = p.parseConditionOnExpr(expr.RHS, right)
			if err != nil {
				return exprList, err
			}
			if expr.Op == influxql.AND {
				exprList = goqu.And(exprList, left)
				exprList = goqu.And(exprList, right)
			} else if expr.Op == influxql.OR {
				exprList = goqu.Or(exprList, left)
				exprList = goqu.Or(exprList, right)
			}
			return exprList, nil
		} else if influxql.EQ <= expr.Op && expr.Op <= influxql.GTE {
			lkey, lok := expr.LHS.(*influxql.VarRef)
			rkey, rok := expr.RHS.(*influxql.VarRef)
			var ok bool
			if lok && !rok {
				exprList, ok, err = p.parseKeyConditionOnExpr(lkey, expr.Op, expr.RHS, exprList)
				if err != nil {
					return exprList, err
				}
				if ok {
					return exprList, nil
				}
			} else if !lok && rok {
				exprList, ok, err = p.parseKeyConditionOnExpr(rkey, reverseOperator(expr.Op), expr.LHS, exprList)
				if err != nil {
					return exprList, err
				}
				if ok {
					return exprList, nil
				}
			}
		}
	case *influxql.ParenExpr:
		return p.parseConditionOnExpr(expr.Expr, exprList)
	}
	exprList, err = p.parseScriptConditionOnExpr(cond, exprList)
	if err != nil {
		return exprList, err
	}
	return exprList, nil
}

func GetColumnLiteral(key string) exp.LiteralExpression {
	return goqu.L(key)
}

func (p *Parser) parseKeyConditionOnExpr(ref *influxql.VarRef, op influxql.Token, val influxql.Expr, exprList exp.ExpressionList) (exp.ExpressionList, bool, error) {
	value, ok, err := getLiteralValue(p.ctx, val)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	column, isNumber := p.ckGetKeyName(ref, influxql.AnyField)

	keyLiteral := GetColumnLiteral(column)

	switch op {
	case influxql.EQ:
		exprList = exprList.Append(keyLiteral.Eq(value))
	case influxql.NEQ:
		exprList = exprList.Append(keyLiteral.Neq(value))
		// neq : is not xxx
		if !isNumber {
			exprList = exprList.Append(goqu.L(fmt.Sprintf("%s != '%s'", column, value)))
		} else {
			exprList = exprList.Append(goqu.L(fmt.Sprintf("%s != %v ", column, value)))
		}
	case influxql.EQREGEX, influxql.NEQREGEX:
		r, ok := value.(*regexp.Regexp)
		if !ok || r == nil {
			return nil, false, fmt.Errorf("invalid regexp '%v'", value)
		}
		reg := strings.Replace(r.String(), `/`, `\/`, -1)
		if op == influxql.EQREGEX {
			exprList = exprList.Append(goqu.L(fmt.Sprintf("extract(%s,'%s') != ''", column, reg)))
		} else {
			exprList = exprList.Append(goqu.L(fmt.Sprintf("extract(%s,'%s') == ''", column, reg)))
		}
	case influxql.LT:
		exprList = exprList.Append(keyLiteral.Lt(value))
	case influxql.LTE:
		exprList = exprList.Append(keyLiteral.Lte(value))
	case influxql.GT:
		exprList = exprList.Append(keyLiteral.Gt(value))
	case influxql.GTE:
		exprList = exprList.Append(keyLiteral.Gte(value))
	default:
		return nil, false, fmt.Errorf("not support operater '%s'", op.String())
	}
	return exprList, true, nil
}

func (p *Parser) ckGetKeyName(ref *influxql.VarRef, deftyp influxql.DataType) (string, bool) {
	name, isNumber, _ := p.ckGetKeyNameAndFlag(ref, deftyp)
	return name, isNumber
}

func (p *Parser) ckGetKeyNameAndFlag(ref *influxql.VarRef, deftyp influxql.DataType) (string, bool, model.ColumnFlag) {
	if newColumn, ok := originColumn[ref.Val]; ok {
		return newColumn, false, model.ColumnFlagNone
	}

	if ref.Type == influxql.Unknown {
		if ref.Val == model.TimestampKey || ref.Val == model.TimeKey {
			return model.TimestampKey, false, model.ColumnFlagTimestamp
		} else if ref.Val == model.NameKey || ref.Val == nameKey {
			return model.NameKey, false, model.ColumnFlagName
		}
		if deftyp == influxql.Tag {
			return ckTagKey(ref.Val), false, model.ColumnFlagTag
		}
	} else if ref.Type == influxql.Tag {
		return ckTagKey(ref.Val), false, model.ColumnFlagTag
	}
	column, isNumber := p.ckFieldKey(ref.Val)
	return column, isNumber, model.ColumnFlagField
}

func (p *Parser) ckColumn(ref *influxql.VarRef) string {
	if newColumn, ok := originColumn[ref.Val]; ok {
		return newColumn
	}

	if ref.Type == influxql.Tag {
		return ckTag(ref.Val)
	}
	column, _ := p.ckField(ref.Val)
	return column
}

func (p *Parser) ckFieldKey(key string) (string, bool) {
	column, isNumber := p.ckField(key)
	if !isNumber {
		return fmt.Sprintf("string_field_values[%s]", column), isNumber
	}
	return fmt.Sprintf("number_field_values[%s]", column), isNumber
}

var originColumn = map[string]string{
	"terminus_key": "tenant_id",
	"org_name":     "org_name",
}

// ckField return clickhouse column, and is number
func (p *Parser) ckField(key string) (string, bool) {
	return fmt.Sprintf("indexOf(number_field_keys,'%s')", key), true
}

func ckTag(key string) string {
	return fmt.Sprintf("indexOf(tag_keys,'%s')", key)
}

func ckTagKey(key string) string {
	return fmt.Sprintf("tag_values[%s]", ckTag(key))
}

func (p *Parser) parseScriptConditionOnExpr(cond influxql.Expr, exprList exp.ExpressionList) (exp.ExpressionList, error) {
	fields := make(map[string]bool)
	s, err := p.getScriptExpressionOnCk(p.ctx, cond, influxql.AnyField, fields)
	if err != nil {
		return exprList, err
	}
	exprList = goqu.And(exprList, goqu.L(s))
	return exprList, nil
}

func (p *Parser) getScriptExpressionOnCk(ctx *Context, expr influxql.Expr, deftyp influxql.DataType, fields map[string]bool) (string, error) {
	if expr == nil {
		return "", nil
	}
	switch expr := expr.(type) {
	case *influxql.BinaryExpr:
		left, err := p.getScriptExpressionOnCk(ctx, expr.LHS, deftyp, fields)
		if err != nil {
			return "", err
		}
		right, err := p.getScriptExpressionOnCk(ctx, expr.RHS, deftyp, fields)
		if err != nil {
			return "", err
		}
		return left + " " + expr.Op.String() + " " + right, nil
	case *influxql.Call:
		val, ok, err := getLiteralFuncValue(ctx, expr)
		if err != nil {
			return "", err
		}
		if ok {
			switch v := val.(type) {
			case nil:
				return "null", nil
			case string:
				return "'" + strings.Replace(v, "'", "\\'", -1) + "'", nil
			case []string:
				return "", fmt.Errorf("not support list in script expression")
			case *regexp.Regexp:
				return "", fmt.Errorf("not support regexp in script expression")
			}
			return fmt.Sprint(val), nil
		}
		s, ok, err := p.getClickhouseFunction(ctx, expr, deftyp, fields)
		if err != nil {
			return "", err
		}
		if ok {
			return s, nil
		}
		return "", fmt.Errorf("not support function '%s' in script expression", expr.Name)
	case *influxql.ParenExpr:
		s, err := p.getScriptExpressionOnCk(ctx, expr.Expr, deftyp, fields)
		if err != nil {
			return "", err
		}
		return "(" + s + ")", nil
	case *influxql.IntegerLiteral:
		return strconv.FormatInt(expr.Val, 10), nil
	case *influxql.NumberLiteral:
		return strconv.FormatFloat(expr.Val, 'f', -1, 64), nil
	case *influxql.BooleanLiteral:
		if expr.Val {
			return strconv.FormatInt(1, 10), nil
		}
		return strconv.FormatInt(0, 10), nil
	case *influxql.UnsignedLiteral:
		return strconv.FormatUint(expr.Val, 10), nil
	case *influxql.StringLiteral:
		return "'" + strings.Replace(expr.Val, "'", "\\'", -1) + "'", nil
	case *influxql.NilLiteral:
		return "null", nil
	case *influxql.VarRef:
		key, _ := p.ckGetKeyName(expr, deftyp)
		return key, nil
	case *influxql.TimeLiteral:
		return strconv.FormatInt(expr.Val.UnixNano(), 10), nil
	case *influxql.DurationLiteral:
		return strconv.FormatInt(int64(expr.Val), 10), nil
	case *influxql.RegexLiteral:
		return "", fmt.Errorf("not support regexp in script expression")
	case *influxql.ListLiteral:
		return "", fmt.Errorf("not support list in script expression")
	}
	return "", fmt.Errorf("invalid expression")
}

func (p *Parser) getClickhouseFunction(ctx *Context, call *influxql.Call, deftyp influxql.DataType, fields map[string]bool) (string, bool, error) {
	fn, ok := ClickHouseFunctions[call.Name]
	if !ok {
		return "", false, nil
	}
	if fn.Convert != nil {
		s, err := fn.Convert(ctx, p, call, deftyp, fields)
		if err != nil {
			return "", false, err
		}
		return s, true, nil
	}
	name, exprs := fn.Name, call.Args
	if fn.Objective {
		if len(call.Args) < 1 {
			return "", false, fmt.Errorf("invalid function")
		}
		obj, err := p.getScriptExpressionOnCk(ctx, call.Args[0], deftyp, fields)
		if err != nil {
			return "", false, err
		}
		name, exprs = "("+obj+")"+"."+fn.Name, call.Args[1:]
	}
	var args []string
	for _, arg := range exprs {
		arg, err := p.getScriptExpressionOnCk(ctx, arg, deftyp, fields)
		if err != nil {
			return "", false, err
		}
		args = append(args, arg)
	}
	return name + "(" + strings.Join(args, ", ") + ")", true, nil
}
