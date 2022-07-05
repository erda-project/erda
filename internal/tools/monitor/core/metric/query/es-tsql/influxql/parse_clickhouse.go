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

const ClickhouseKind = "Clickhouse"

func (p *Parser) ParseClickhouse(s *influxql.SelectStatement) (tsql.Query, error) {

	sources, err := p.from(s.Sources)

	if err != nil {
		return nil, fmt.Errorf("select stmt parse to from is error: %v", err)
	}

	expr := goqu.From("metric") // metric is fake, in execution layer it's real table
	p.appendTimeKeyByExpr(expr)

	// add parser filter to expr
	expr, err = p.filterOnExpr(expr)
	if err != nil {
		return nil, err
	}

	expr, err = p.conditionOnExpr(expr, s)
	if err != nil {
		return nil, err
	}

	// select
	expr, handlers, columns, err := p.parseQueryOnExpr(s.Fields, expr)
	if err != nil {
		return nil, err
	}

	expr, tailLiters, err := p.ParseGroupByOnExpr(s.Dimensions, expr, handlers, columns)
	if err != nil {
		return nil, err
	}

	expr, err = p.ParseOrderByOnExpr(s.SortFields, expr, columns)
	if err != nil {
		return nil, err
	}

	expr, err = p.ParseOffsetAndLimitOnExpr(s, expr)
	if err != nil {
		return nil, err
	}
	return QueryClickhouse{
		sources:   sources,
		subLiters: tailLiters,
		column:    handlers,
		start:     p.ctx.start,
		end:       p.ctx.end,
		kind:      ClickhouseKind,
		ctx:       p.ctx,
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
				sortFields[ckGetKeyName(v, influxql.AnyField)] = field.Ascending
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
	if s.Limit <= 0 {
		s.Limit = model.DefaultLimtSize
	}
	if s.Offset <= 0 {
		s.Offset = 0
	}
	expr = expr.Offset(uint(s.Offset))
	expr = expr.Limit(uint(s.Limit))
	return expr, nil
}

func (p *Parser) ParseGroupByOnExpr(dimensions influxql.Dimensions, expr *goqu.SelectDataset, handlers []*SQLColumnHandler, columns map[string]string) (*goqu.SelectDataset, map[string]string, error) {
	if len(dimensions) <= 0 {
		return expr, nil, nil
	}
	expr, liters, tailLiters, err := p.parseQueryDimensionsByExpr(expr, dimensions, handlers)
	if err != nil {
		return nil, nil, err
	}

	for _, script := range liters {
		c := script
		if newName, ok := columns[script]; ok {
			c = newName
		}
		// simple column
		expr = expr.GroupByAppend(goqu.C(c))
		continue
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
		expr = expr.SelectAppend(goqu.L(column).As(asName))
	}
	return expr, handlers, columns, nil
}
func (p *Parser) parseQueryDimensionsByExpr(exprSelect *goqu.SelectDataset, dimensions influxql.Dimensions, handler []*SQLColumnHandler) (*goqu.SelectDataset, []string, map[string]string, error) {
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
				exprSelect = exprSelect.SelectAppend(goqu.L(fmt.Sprintf("toInt64(toStartOfInterval(timestamp, toIntervalSecond(%v)))", interval)).As(p.ctx.TimeKey()))

				// ORDER BY timestamp with fill from 0 to 0 step 0
				// todo goqu order statement should be literal + asc, but with fill is order by `column` [asc/desc] with fill from %left to %right step %interval
				tailExpr[p.ctx.TimeKey()] = fmt.Sprintf("%s with fill from %v to %v step %v", p.ctx.TimeKey(), start, end, interval)

				exprList = append(exprList, p.ctx.TimeKey())

				handler = append(handler, &SQLColumnHandler{
					field: &influxql.Field{
						Expr:  expr,
						Alias: p.ctx.TimeKey(),
					},
				})

				continue
			} else if expr.Name == "range" {
				if len(tailExpr) > 0 {
					return nil, nil, nil, fmt.Errorf("with fill statement should be one function")
				}

				// GROUP BY range(plt::field, 0,8000,1000) field, min, max, step
				err := mustCallArgsMinNum(expr, 3)
				if err != nil {
					return exprSelect, nil, nil, err
				}

				field, ok := expr.Args[0].(*influxql.VarRef)
				if !ok {
					return exprSelect, nil, nil, fmt.Errorf("args[0] is not reference in 'range' function")
				}

				//
				min, ok := expr.Args[1].(*influxql.IntegerLiteral)
				if !ok {
					return exprSelect, nil, nil, fmt.Errorf("args[1] is not reference in 'range' function")
				}

				max, ok := expr.Args[2].(*influxql.IntegerLiteral)
				if !ok {
					return exprSelect, nil, nil, fmt.Errorf("args[2] is not reference in 'range' function")
				}

				step, ok := expr.Args[3].(*influxql.IntegerLiteral)
				if !ok {
					return exprSelect, nil, nil, fmt.Errorf("args[3] is not reference in 'range' function")
				}

				tailExpr["range"] = fmt.Sprintf("order by %v with fill from %v to %v step %v",
					ckGetKeyName(field, influxql.AnyField), min.Val, max.Val, step.Val)

				// add select
				exprSelect = exprSelect.SelectAppend(goqu.L(fmt.Sprintf("intDiv(%s,%v) * %v", ckGetKeyName(field, influxql.AnyField), step, step)).As("range"))

				// add group by
				exprList = append(exprList, "range")

				// add select handler
				handler = append(handler, &SQLColumnHandler{
					field: &influxql.Field{
						Expr:  expr,
						Alias: "range",
					},
				})

				continue
			}

			script, err := getScriptExpression(p.ctx, dim.Expr, influxql.Tag, nil)
			if err != nil {
				return exprSelect, nil, nil, nil
			}
			exprList = append(exprList, script)
		}
		script := getExprStringAndFlagByExpr(dim.Expr, influxql.Tag)
		exprList = append(exprList, script)
	}
	return exprSelect, exprList, tailExpr, nil
}

func getExprStringAndFlagByExpr(expr influxql.Expr, deftyp influxql.DataType) (key string) {
	if expr == nil {
		return ""
	}
	switch expr := expr.(type) {
	case *influxql.BinaryExpr:
		left := getExprStringAndFlagByExpr(expr.LHS, deftyp)
		right := getExprStringAndFlagByExpr(expr.RHS, deftyp)
		return left + expr.Op.String() + right
	case *influxql.Call:
		var args []string
		for _, arg := range expr.Args {
			k := getExprStringAndFlagByExpr(arg, deftyp)
			args = append(args, k)
		}
		return expr.Name + "(" + strings.Join(args, ",") + ")"
	case *influxql.ParenExpr:
		key = getExprStringAndFlagByExpr(expr.Expr, deftyp)
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
		key, _ = ckGetKeyNameAndFlag(expr, deftyp)
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
	if ch.AllColumns() {
		return ch, nil
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
				err = fn.Aggregations(aggs, FuncFlagSelect)
				if err != nil {
					return err
				}
				fns[id] = fn
			}
		} else if _, ok := tsql.BuildInFunctions[expr.Name]; ok {
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
		_, ok := AggFunctions[expr.Name]
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
		cols[ckGetKeyName(expr, influxql.AnyField)] = expr.Val
	}
	return nil
}

func (p *Parser) appendTimeKeyByExpr(expr *goqu.SelectDataset) {
	start, end := p.ctx.Range(true)
	expr = expr.Where(
		goqu.C(p.ctx.timeKey).Gte(start),
		goqu.C(p.ctx.timeKey).Lt(end),
	)
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

func filterToExpr(filters []*model.Filter, expr *goqu.SelectDataset) (*goqu.SelectDataset, error) {
	or := goqu.Or()
	expressionList := goqu.And()

	for _, item := range filters {
		switch item.Operator {
		case "eq", "=", "":
			expressionList = expressionList.Append(goqu.C(item.Key).Eq(item.Value))
		case "neq", "!=":
			expressionList = expressionList.Append(goqu.C(item.Key).Neq(item.Value))
		case "gt", ">":
			expressionList = expressionList.Append(goqu.C(item.Key).Gt(item.Value))
		case "gte", ">=":
			expressionList = expressionList.Append(goqu.C(item.Key).Gte(item.Value))
		case "lt", "<":
			expressionList = expressionList.Append(goqu.C(item.Key).Lt(item.Value))
		case "lte", "<=":
			expressionList = expressionList.Append(goqu.C(item.Key).Lte(item.Value))
		case "in":
			if values, ok := item.Value.([]interface{}); ok {
				expressionList = expressionList.Append(goqu.C(item.Key).In(values))
			}
		case "match":
			expressionList = expressionList.Append(goqu.C(item.Key).Like(fmt.Sprintf("%%%v%%", item.Value)))
		case "nmatch":
			expressionList = expressionList.Append(goqu.C(item.Key).NotLike(fmt.Sprintf("%%%v%%", item.Value)))

		case "or_eq":
			orExpr := goqu.C(item.Key).Eq(item.Value)

			if or.IsEmpty() {
				or = goqu.Or(orExpr)
			} else {
				or = or.Append(orExpr)
			}

		case "or_in":
			if values, ok := item.Value.([]interface{}); ok {
				orExpr := goqu.C(item.Key).In(values)
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
		expr, err = filterToExpr(p.filter, expr)
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
	keyLiteral := GetColumnLiteral(ckGetKeyName(ref, influxql.Tag))

	switch op {
	case influxql.EQ:
		exprList = exprList.Append(keyLiteral.Eq(value))
	case influxql.NEQ:
		exprList = exprList.Append(keyLiteral.Neq(value))
	case influxql.EQREGEX, influxql.NEQREGEX:
		r, ok := value.(*regexp.Regexp)
		if !ok || r == nil {
			return nil, false, fmt.Errorf("invalid regexp '%v'", value)
		}
		reg := strings.Replace(r.String(), `/`, `\/`, -1)
		if op == influxql.EQREGEX {
			exprList = exprList.Append(keyLiteral.RegexpLike(reg))
		} else {
			exprList = exprList.Append(keyLiteral.RegexpNotLike(reg))
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

func ckGetKeyName(ref *influxql.VarRef, deftyp influxql.DataType) string {
	name, _ := ckGetKeyNameAndFlag(ref, deftyp)
	return name
}

func ckGetKeyNameAndFlag(ref *influxql.VarRef, deftyp influxql.DataType) (string, model.ColumnFlag) {
	if ref.Type == influxql.Unknown {
		if ref.Val == model.TimestampKey || ref.Val == model.TimeKey {
			return model.TimestampKey, model.ColumnFlagTimestamp
		} else if ref.Val == model.NameKey || ref.Val == nameKey {
			return model.NameKey, model.ColumnFlagName
		}
		if deftyp == influxql.Tag {
			return ckTagKey(ref.Val), model.ColumnFlagTag
		}
	} else if ref.Type == influxql.Tag {
		return ckTagKey(ref.Val), model.ColumnFlagTag
	}
	return ckFieldKey(ref.Val), model.ColumnFlagField
}

func ckFieldKey(key string) string {
	// 	return fmt.Sprintf("string_field_values[string_field_keys(tag_keys,'%s')]", key)
	return fmt.Sprintf("number_field_values[indexOf(number_field_keys,'%s')]", key)
}

func ckTagKey(key string) string {
	return fmt.Sprintf("tag_values[indexOf(tag_keys,'%s')]", key)
}

func (p *Parser) parseScriptConditionOnExpr(cond influxql.Expr, exprList exp.ExpressionList) (exp.ExpressionList, error) {
	fields := make(map[string]bool)
	s, err := getScriptExpressionOnCk(p.ctx, cond, influxql.Tag, fields)
	if err != nil {
		return exprList, err
	}
	exprList = goqu.And(exprList, goqu.L(s))
	return exprList, nil
}

func getScriptExpressionOnCk(ctx *Context, expr influxql.Expr, deftyp influxql.DataType, fields map[string]bool) (string, error) {
	if expr == nil {
		return "", nil
	}
	switch expr := expr.(type) {
	case *influxql.BinaryExpr:
		left, err := getScriptExpressionOnCk(ctx, expr.LHS, deftyp, fields)
		if err != nil {
			return "", err
		}
		right, err := getScriptExpressionOnCk(ctx, expr.RHS, deftyp, fields)
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
		s, ok, err := getClickhouseFuntion(ctx, expr, deftyp, fields)
		if err != nil {
			return "", err
		}
		if ok {
			return s, nil
		}
		return "", fmt.Errorf("not support function '%s' in script expression", expr.Name)
	case *influxql.ParenExpr:
		s, err := getScriptExpression(ctx, expr.Expr, deftyp, fields)
		if err != nil {
			return "", err
		}
		return "(" + s + ")", nil
	case *influxql.IntegerLiteral:
		return strconv.FormatInt(expr.Val, 10), nil
	case *influxql.NumberLiteral:
		return strconv.FormatFloat(expr.Val, 'f', -1, 64), nil
	case *influxql.BooleanLiteral:
		return strconv.FormatBool(expr.Val), nil
	case *influxql.UnsignedLiteral:
		return strconv.FormatUint(expr.Val, 10), nil
	case *influxql.StringLiteral:
		return "'" + strings.Replace(expr.Val, "'", "\\'", -1) + "'", nil
	case *influxql.NilLiteral:
		return "null", nil
	case *influxql.VarRef:
		key := ckGetKeyName(expr, deftyp)
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

func getClickhouseFuntion(ctx *Context, call *influxql.Call, deftyp influxql.DataType, fields map[string]bool) (string, bool, error) {
	fn, ok := ClickHouseFunctions[call.Name]
	if !ok {
		return "", false, nil
	}
	if fn.Convert != nil {
		s, err := fn.Convert(ctx, call, deftyp, fields)
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
		obj, err := getScriptExpressionOnCk(ctx, call.Args[0], deftyp, fields)
		if err != nil {
			return "", false, err
		}
		name, exprs = "("+obj+")"+"."+fn.Name, call.Args[1:]
	}
	var args []string
	for _, arg := range exprs {
		arg, err := getScriptExpressionOnCk(ctx, arg, deftyp, fields)
		if err != nil {
			return "", false, err
		}
		args = append(args, arg)
	}
	return name + "(" + strings.Join(args, ", ") + ")", true, nil
}
