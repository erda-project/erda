package esinfluxql

import (
	"fmt"

	"github.com/influxdata/influxql"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
)

type _column struct {
	key                string
	asName             string
	rootColumn         string
	isWildcard         bool
	isNoArrayKey       bool
	isTimeKey          bool
	isNumberField      bool
	isStringField      bool
	isAggFunctionField bool
	expr               string
	flag               model.ColumnFlag
}

func (c _column) modelKey() string {
	if c.isWildcard || c.isNoArrayKey {
		return ""
	} else if c.isStringField {
		return "string_field_keys"
	} else if c.isStringField {
		return "number_field_keys"
	}
	return "tag_keys"
}

type _columns struct {
	columns map[string]_column
}

func newColumns() *_columns {
	return &_columns{
		columns: make(map[string]_column),
	}
}
func (c *_columns) addColumn(key string, column _column) {
	column.key = key
	c.columns[key] = column
}

func (c *_columns) getColumn(key string) (_column, bool) {
	v, ok := c.columns[key]
	return v, ok
}

func (c *_columns) addCallColumn(expr *influxql.Call, key string) {
	c.addColumn(key, _column{
		expr: expr.String(),
	})
}

func (c *_columns) addDimensionColumn(expr influxql.Expr, key string) {
	c.addColumn(key, _column{
		expr:   expr.String(),
		asName: expr.String(),
	})
}

func (c *_columns) addTimeBucketColumn(timeKey string, intervalSeconds int64) {
	timeBucketColumn := fmt.Sprintf("bucket_%s", timeKey)
	timeBucketExpr := fmt.Sprintf("intDiv(toRelativeSecondNum(timestamp), %v)", intervalSeconds)

	_column := _column{
		expr:       timeBucketExpr,
		asName:     timeBucketColumn,
		rootColumn: timeBucketColumn,
		isTimeKey:  true,
	}
	c.addColumn(timeBucketColumn, _column)

}

func (c *_columns) addWildcard() {
	c.addColumn("*", _column{asName: "*", isWildcard: true, rootColumn: "*"})
}
