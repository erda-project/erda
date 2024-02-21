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
	"strings"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/influxdata/influxql"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric/model"
	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

func TestClickhouse(t *testing.T) {
	tests := []struct {
		name    string
		stm     string
		params  map[string]interface{}
		require func(t *testing.T, queries []tsql.Query)
	}{
		{
			name: "select * from metrics",
			stm:  "select * from metrics",
			require: func(t *testing.T, queries []tsql.Query) {

			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := New(0, 0, test.stm, false)
			if test.params != nil {
				p.SetParams(test.params)
			}
			err := p.Build()
			require.NoError(t, err)
			queries, err := p.ParseQuery(context.Background(), model.ClickhouseKind)
			require.NoError(t, err)
			test.require(t, queries)
		})
	}

}

func TestFilterToExpr(t *testing.T) {
	tests := []struct {
		name string
		args []*model.Filter
		want string
	}{
		{
			name: "empty",
			args: nil,
			want: "SELECT * FROM \"table\"",
		},
		{
			name: "eq",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "123",
					Operator: "=",
				},
				{
					Key:      "column",
					Value:    "123",
					Operator: "eq",
				},
				{
					Key:      "column",
					Value:    "123",
					Operator: "",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column = '123') AND (column = '123') AND (column = '123'))",
		},
		{
			name: "neq",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "123",
					Operator: "neq",
				},
				{
					Key:      "column",
					Value:    "123",
					Operator: "!=",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column != '123') AND (column != '123'))",
		},
		{
			name: "gt",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "123",
					Operator: "gt",
				},
				{
					Key:      "column",
					Value:    "123",
					Operator: ">",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column > '123') AND (column > '123'))",
		},
		{
			name: "gte",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "123",
					Operator: "gte",
				},
				{
					Key:      "column",
					Value:    "123",
					Operator: ">=",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column >= '123') AND (column >= '123'))",
		},
		{
			name: "lt",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "123",
					Operator: "lt",
				},
				{
					Key:      "column",
					Value:    "123",
					Operator: "<",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column < '123') AND (column < '123'))",
		},
		{
			name: "lte",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "123",
					Operator: "lte",
				},
				{
					Key:      "column",
					Value:    "123",
					Operator: "<=",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column <= '123') AND (column <= '123'))",
		},
		{
			name: "in",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "",
					Operator: "in",
				},
				{
					Key:      "column",
					Value:    []interface{}{"1111", "2222", "3333"},
					Operator: "in",
				},
			},
			want: "SELECT * FROM \"table\" WHERE (column IN ('1111', '2222', '3333'))",
		},
		{
			name: "match",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "12313",
					Operator: "match",
				},
			},
			want: "SELECT * FROM \"table\" WHERE (column LIKE '%12313%')",
		},
		{
			name: "match *",
			args: []*model.Filter{
				{
					Key:      "tags.addon_id",
					Value:    "*",
					Operator: "match",
				},
			},
			want: "SELECT * FROM \"table\"",
		},
		{
			name: "match * content",
			args: []*model.Filter{
				{
					Key:      "tags.addon_id",
					Value:    "*123",
					Operator: "match",
				},
			},
			want: "SELECT * FROM \"table\" WHERE (tag_values[indexOf(tag_keys,'addon_id')] LIKE '%%123%')",
		},
		{
			name: "nmatch",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "12313",
					Operator: "nmatch",
				},
			},
			want: "SELECT * FROM \"table\" WHERE (column NOT LIKE '%12313%')",
		},
		{
			name: "nmatch *",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "*",
					Operator: "nmatch",
				},
			},
			want: "SELECT * FROM \"table\"",
		},
		{
			name: "nmatch * content",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "12313*",
					Operator: "nmatch",
				},
			},
			want: "SELECT * FROM \"table\" WHERE (column NOT LIKE '%12313%%')",
		},
		{
			name: "or eq",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "12313",
					Operator: "eq",
				},
				{
					Key:      "column",
					Value:    "12313",
					Operator: "or_eq",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column = '12313') AND (column = '12313'))",
		},
		{
			name: "or in",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    []interface{}{"111", "222", "3333"},
					Operator: "or_in",
				},
				{
					Key:      "column",
					Value:    []interface{}{"111", "222", "3333"},
					Operator: "in",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((column IN ('111', '222', '3333')) AND (column IN ('111', '222', '3333')))",
		},
		{
			name: "and or and or",
			args: []*model.Filter{
				{
					Key:      "column",
					Value:    "12313",
					Operator: "eq",
				},
				{
					Key:      "column",
					Value:    "12313",
					Operator: "or_eq",
				},
				{
					Key:      "column",
					Value:    "12313",
					Operator: "eq",
				},
				{
					Key:      "column",
					Value:    "12313",
					Operator: "or_eq",
				},
			},
			want: "SELECT * FROM \"table\" WHERE (((column = '12313') AND (column = '12313')) AND ((column = '12313') OR (column = '12313')))",
		},
		{
			name: "tags filter",
			args: []*model.Filter{
				{
					Key:      "tags.cluster_name",
					Value:    "123",
					Operator: "eq",
				},
				{
					Key:      "tags.org_name",
					Value:    "123",
					Operator: "eq",
				},
			},
			want: "SELECT * FROM \"table\" WHERE ((tag_values[indexOf(tag_keys,'cluster_name')] = '123') AND (org_name = '123'))",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr := goqu.From("table")

			p := Parser{}

			exprList := goqu.And()
			exprList, err := p.filterToExpr(test.args, exprList)
			require.NoError(t, err)
			expr = expr.Where(exprList)
			sql, _, err := expr.ToSQL()
			require.NoError(t, err)
			require.Equal(t, test.want, sql)
		})
	}
}

func TestConditionToExpr(t *testing.T) {
	tests := []struct {
		name      string
		originSql string
		wantError bool
		want      string
		params    map[string]interface{}
	}{
		{
			name:      "all and",
			originSql: "select * from table where column::tag = 1 and column::tag = 2 and column::tag = 3",
			want:      "SELECT * FROM \"table\" WHERE (((tag_values[indexOf(tag_keys,'column')] = 1) AND (tag_values[indexOf(tag_keys,'column')] = 2)) AND (tag_values[indexOf(tag_keys,'column')] = 3))",
		},
		{
			name:      "all or",
			originSql: "select * from table where column::tag = 1 or column::tag = 2 or column::tag = 3",
			want:      "SELECT * FROM \"table\" WHERE (((tag_values[indexOf(tag_keys,'column')] = 1) OR (tag_values[indexOf(tag_keys,'column')] = 2)) OR (tag_values[indexOf(tag_keys,'column')] = 3))",
		},
		{
			name:      "and or (and)or",
			originSql: "select * from table where (column::tag = 1 and column::tag = 2) or column::tag = 3",
			want:      "SELECT * FROM \"table\" WHERE (((tag_values[indexOf(tag_keys,'column')] = 1) AND (tag_values[indexOf(tag_keys,'column')] = 2)) OR (tag_values[indexOf(tag_keys,'column')] = 3))",
		},
		{
			name:      "and or and(or)",
			originSql: "select * from table where column::tag = 1 and (column::tag = 2 or column::tag = 3)",
			want:      "SELECT * FROM \"table\" WHERE ((tag_values[indexOf(tag_keys,'column')] = 1) AND ((tag_values[indexOf(tag_keys,'column')] = 2) OR (tag_values[indexOf(tag_keys,'column')] = 3)))",
		},
		{
			name:      "include function",
			originSql: "select * from table where include(target_service_id::tag, '2673_feature/test1_apm-demo-api','2673_feature/test1_apm-demo-dubbo','2673_feature/test1_apm-demo-ui')",
			params: map[string]interface{}{
				"terminus_key": "111",
			},
			want: "SELECT * FROM \"table\" WHERE ((tag_values[indexOf(tag_keys,'target_service_id')])=('2673_feature/test1_apm-demo-api') or (tag_values[indexOf(tag_keys,'target_service_id')])=('2673_feature/test1_apm-demo-dubbo') or (tag_values[indexOf(tag_keys,'target_service_id')])=('2673_feature/test1_apm-demo-ui'))",
		},
		{
			name:      "key gt,lt",
			originSql: "select * from table where column::tag > column2::tag  and (column::tag > 2 or column::tag < 3)",
			want:      "SELECT * FROM \"table\" WHERE (tag_values[indexOf(tag_keys,'column')] > tag_values[indexOf(tag_keys,'column2')] AND ((tag_values[indexOf(tag_keys,'column')] > 2) OR (tag_values[indexOf(tag_keys,'column')] < 3)))",
		},
		{
			name:      "key gte,lte",
			originSql: "select * from table where column::tag >= column2::tag  and (column::tag >= 2 or column::tag <= 3)",
			want:      "SELECT * FROM \"table\" WHERE (tag_values[indexOf(tag_keys,'column')] >= tag_values[indexOf(tag_keys,'column2')] AND ((tag_values[indexOf(tag_keys,'column')] >= 2) OR (tag_values[indexOf(tag_keys,'column')] <= 3)))",
		},
		{
			name:      "regex",
			originSql: "select * from table where column::tag =~/cpu[0-2]$/",
			want:      "SELECT * FROM \"table\" WHERE extract(tag_values[indexOf(tag_keys,'column')],'cpu[0-2]$') != ''",
		},
		{
			name:      "not equal",
			originSql: "select * from table where column::tag = 123 and column::tag != 123",
			want:      "SELECT * FROM \"table\" WHERE ((tag_values[indexOf(tag_keys,'column')] = 123) AND ((tag_values[indexOf(tag_keys,'column')] != 123) AND tag_values[indexOf(tag_keys,'column')] != '%!s(int64=123)'))",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Parser{}
			parse := influxql.NewParser(strings.NewReader(test.originSql))
			if len(test.params) > 0 {
				parse.SetParams(test.params)
			}
			q, err := parse.ParseQuery()

			require.NoError(t, err)
			selectStmt, ok := q.Statements[0].(*influxql.SelectStatement)

			expr := goqu.From("table")
			require.Truef(t, ok, "parse query is not select statement")
			expr, err = p.conditionOnExpr(expr, selectStmt)
			if test.wantError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			sql, _, err := expr.ToSQL()
			require.NoError(t, err)
			require.Equal(t, test.want, sql)
		})
	}
}

func TestSelect(t *testing.T) {
	tests := []struct {
		name   string
		sql    string
		want   string
		params map[string]interface{}
	}{
		{
			name: "select tenant_id",
			sql:  "select tenant_id from table",
			want: "SELECT toNullable(tenant_id) AS \"tenant_id\" FROM \"table\"",
		},
		{
			name: "select column",
			sql:  "select column from table",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column\" FROM \"table\"",
		},
		{
			name: "select sum(column)",
			sql:  "select sum(column) from table",
			want: "SELECT SUM(number_field_values[indexOf(number_field_keys,'column')]) AS \"64fa1afe5a1e10fc\" FROM \"table\"",
		},
		{
			name: "select sum",
			sql:  "select sum(column)from table",
			want: "SELECT SUM(number_field_values[indexOf(number_field_keys,'column')]) AS \"64fa1afe5a1e10fc\" FROM \"table\"",
		},
		{
			name: "select distinct",
			sql:  "select DISTINCT(service_id::tag) from table",
			want: "SELECT uniqCombined(if(indexOf(tag_keys,'service_id') == 0,null,tag_values[indexOf(tag_keys,'service_id')])) AS \"2e1cf76dc42cffe8\" FROM \"table\"",
		},
		{
			name: "select origin distinct",
			sql:  "select DISTINCT(timestamp) from table",
			want: "SELECT uniqCombined(timestamp) AS \"05b05e5af080d601\" FROM \"table\"",
		},
		{
			name: "select max",
			sql:  "select max(column) from table",
			want: "SELECT MAX(if(indexOf(number_field_keys,'column') == 0,null,number_field_values[indexOf(number_field_keys,'column')])) AS \"322cc30ad1d92b84\" FROM \"table\"",
		},
		{
			name: "select min",
			sql:  "select min(service_id::tag) from table",
			want: "SELECT MIN(if(indexOf(tag_keys,'service_id') == 0,null,tag_values[indexOf(tag_keys,'service_id')])) AS \"b784dcb694669c75\" FROM \"table\"",
		},
		{
			name: "select origin min",
			sql:  "select min(timestamp) from table",
			want: "SELECT MIN(timestamp) AS \"5636e73fd0aba08e\" FROM \"table\"",
		},
		{
			name: "select count",
			sql:  "select count(service_id::tag) from table",
			want: "SELECT COUNT(if(indexOf(tag_keys,'service_id') == 0,null,tag_values[indexOf(tag_keys,'service_id')])) AS \"993567cffecb105e\" FROM \"table\"",
		},
		{
			name: "select sum(if:eq)",
			sql:  "select sum(if(eq(error::tag, 'true'),elapsed_count::field,0)) from table",
			want: "SELECT SUM(if(tag_values[indexOf(tag_keys,'error')]='true',number_field_values[indexOf(number_field_keys,'elapsed_count')],0)) AS \"72e4961c054d5bb4\" FROM \"table\"",
		},
		{
			name: "select sum(if:gt)",
			sql:  "select sum(if(gt(error::tag, 100),elapsed_count::field,0)) from table",
			want: "SELECT SUM(if(tag_values[indexOf(tag_keys,'error')]>100,number_field_values[indexOf(number_field_keys,'elapsed_count')],0)) AS \"050269978f5ba682\" FROM \"table\"",
		},
		{
			name: "select sum(if:include)",
			sql:  "select sum(if(include(error::tag, '11','22','33'),elapsed_count::field,0)) from table",
			want: "SELECT SUM(if(((tag_values[indexOf(tag_keys,'error')])=('11') or (tag_values[indexOf(tag_keys,'error')])=('22') or (tag_values[indexOf(tag_keys,'error')])=('33')),number_field_values[indexOf(number_field_keys,'elapsed_count')],0)) AS \"19b1b14fbb96ec2c\" FROM \"table\"",
		},
		{
			name: "select sum(if:not_include)",
			sql:  "select sum(if(not_include(error::tag, '11','22','33'),elapsed_count::field,0)) from table",
			want: "SELECT SUM(if(((tag_values[indexOf(tag_keys,'error')])!=('11') and (tag_values[indexOf(tag_keys,'error')])!=('22') and (tag_values[indexOf(tag_keys,'error')])!=('33')),number_field_values[indexOf(number_field_keys,'elapsed_count')],0)) AS \"c489eef269fb2d01\" FROM \"table\"",
		},
		{
			name: "select round_float(avg)",
			sql:  "select round_float(avg(committed::field), 2) from table",
			want: "SELECT AVG(if(indexOf(number_field_keys,'committed') == 0,null,number_field_values[indexOf(number_field_keys,'committed')])) AS \"cd848468318e898b\" FROM \"table\"",
		},
		{
			name: "select time diff function",
			sql:  "SELECT service_instance_id::tag,if(gt(now()-max(timestamp),300000000000),'false','true') AS state from table",
			want: "SELECT MAX(timestamp) AS \"1362043e612fc3f5\", toNullable(tag_values[indexOf(tag_keys,'service_instance_id')]) AS \"service_instance_id::tag\" FROM \"table\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Parser{
				ctx: &Context{},
			}
			parse := influxql.NewParser(strings.NewReader(test.sql))
			if len(test.params) > 0 {
				parse.SetParams(test.params)
			}
			q, err := parse.ParseQuery()

			require.NoError(t, err)
			selectStmt, ok := q.Statements[0].(*influxql.SelectStatement)

			expr := goqu.From("table")
			require.Truef(t, ok, "parse query is not select statement")
			expr, handler, _, err := p.parseQueryOnExpr(selectStmt.Fields, expr)

			_ = handler

			require.NoError(t, err)

			sql, _, err := expr.ToSQL()

			require.NoError(t, err)
			require.Equal(t, test.want, sql)
		})
	}
}

func TestGroupBy(t *testing.T) {
	tests := []struct {
		name   string
		sql    string
		want   string
		params map[string]interface{}
	}{
		{
			name: "count,column",
			sql:  "select column,count(1) from table group by column",
			want: "SELECT COUNT(1) AS \"1c2fcd7a03c386f7\", number_field_values[indexOf(number_field_keys,'column')] AS \"column\" FROM \"table\" GROUP BY \"column\"",
		},
		{
			name: "time(),max(column)",
			sql:  "select max(column) from table group by time()",
			want: "SELECT MAX(if(indexOf(tag_keys,'column') == 0,null,tag_values[indexOf(tag_keys,'column')])) AS \"322cc30ad1d92b84\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY intDiv(toRelativeSecondNum(timestamp), 60)",
		},
		{
			name: "group_not_select_column",
			sql:  "select column from table group by max(column),column",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column\" FROM \"table\"",
		},
		{
			name: "time(2h)",
			sql:  "select column from table group by time(2h)",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"column\", intDiv(toRelativeSecondNum(timestamp), 7200)",
		},
		{
			name: "groupby,time()",
			sql:  "select sum(http_status_code_count::field),http_status_code::tag from table GROUP BY time(),http_status_code::tag",
			want: "SELECT SUM(number_field_values[indexOf(number_field_keys,'http_status_code_count')]) AS \"339c9df3d700c4f0\", toNullable(tag_values[indexOf(tag_keys,'http_status_code')]) AS \"http_status_code::tag\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"http_status_code::tag\", intDiv(toRelativeSecondNum(timestamp), 60)",
		},
		{
			name: "time(),column",
			sql:  "select http_status_code::tag from table GROUP BY time()",
			want: "SELECT toNullable(tag_values[indexOf(tag_keys,'http_status_code')]) AS \"http_status_code::tag\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"http_status_code::tag\", intDiv(toRelativeSecondNum(timestamp), 60)",
		},
		{
			name: "no group",
			sql:  "select max(http_status_code::tag) from table",
			want: "SELECT MAX(if(indexOf(tag_keys,'http_status_code') == 0,null,tag_values[indexOf(tag_keys,'http_status_code')])) AS \"61421335fd474c8e\" FROM \"table\"",
		},
		{
			name: "group tostring function",
			sql:  "select max(http_status_code::tag),tostring(timestamp) from table group by tostring(timestamp)",
			want: "SELECT MAX(if(indexOf(tag_keys,'http_status_code') == 0,null,tag_values[indexOf(tag_keys,'http_status_code')])) AS \"61421335fd474c8e\", toNullable(timestamp) AS \"timestamp\" FROM \"table\" GROUP BY \"timestamp\", '', tostring(timestamp)",
		},
		{
			name: "group sub function",
			sql:  "SELECT service_instance_id::tag,if(gt(now()-timestamp,300000000000),'false','true') as state from table group by time()",
			want: "SELECT toNullable(tag_values[indexOf(tag_keys,'service_instance_id')]) AS \"service_instance_id::tag\", toNullable(timestamp) AS \"timestamp\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"timestamp\", intDiv(toRelativeSecondNum(timestamp), 60), \"service_instance_id::tag\"",
		},
		{
			name: "group by name tag",
			sql:  "SELECT name::tag from table group by time(),name::tag",
			want: "SELECT toNullable(tag_values[indexOf(tag_keys,'name')]) AS \"name::tag\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"name::tag\", intDiv(toRelativeSecondNum(timestamp), 60)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Parser{
				ctx: &Context{
					originalTimeUnit: tsql.Nanosecond,
					targetTimeUnit:   tsql.Millisecond,
					dimensions:       make(map[string]bool),
				},
			}
			parse := influxql.NewParser(strings.NewReader(test.sql))
			if len(test.params) > 0 {
				parse.SetParams(test.params)
			}
			q, err := parse.ParseQuery()

			require.NoError(t, err)
			selectStmt, ok := q.Statements[0].(*influxql.SelectStatement)

			expr := goqu.From("table")
			require.Truef(t, ok, "parse query is not select statement")

			expr, handler, columns, err := p.parseQueryOnExpr(selectStmt.Fields, expr)

			expr, _, err = p.ParseGroupByOnExpr(selectStmt.Dimensions, expr, &handler, columns)

			require.NoError(t, err)

			sql, _, err := expr.ToSQL()
			require.NoError(t, err)

			if strings.Index(sql, "ORDER BY") != -1 {
				sql = sql[:strings.Index(sql, "ORDER BY")]
			}

			t.Log("arg sql", test.sql)
			t.Log("got", sql)
			t.Log("want", test.want)

			if strings.Index(test.want, strings.ToUpper("group by")) <= 0 {
				require.Equal(t, test.want, sql)
				return
			}
			sqlGroup := strings.Split(sql, strings.ToUpper("group by"))[1]
			wantGroup := strings.Split(test.want, strings.ToUpper("group by"))[1]

			gotGroupArr := strings.Split(strings.ReplaceAll(sqlGroup, " ", ""), ",")
			wantGroupArr := strings.Split(strings.ReplaceAll(wantGroup, " ", ""), ",")
			require.ElementsMatch(t, wantGroupArr, gotGroupArr)
		})
	}
}

func TestOffsetLimit(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		want    string
		wantErr bool
	}{
		{
			name: "limit",
			sql:  "select * from table limit 10",
			want: "SELECT * FROM \"table\" LIMIT 10",
		},
		{
			name: "offset",
			sql:  "select * from table offset 100",
			want: "SELECT * FROM \"table\" OFFSET 100", // offset zero is ignore, limit model.DefaultLimtSize
		},
		{
			name: "limit,offset",
			sql:  "select * from table limit 10 offset 30 ",
			want: "SELECT * FROM \"table\" LIMIT 10 OFFSET 30",
		},
		{
			name: "none",
			sql:  "SELECT * FROM table",
			want: "SELECT * FROM \"table\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			p := Parser{
				ctx: &Context{},
			}
			parse := influxql.NewParser(strings.NewReader(test.sql))

			q, err := parse.ParseQuery()

			require.NoError(t, err)
			selectStmt, ok := q.Statements[0].(*influxql.SelectStatement)

			expr := goqu.From("table")
			require.Truef(t, ok, "parse query is not select statement")

			expr, err = p.ParseOffsetAndLimitOnExpr(selectStmt, expr)

			require.NoError(t, err)

			sql, _, err := expr.ToSQL()

			require.NoError(t, err)
			require.Equal(t, test.want, sql)

		})
	}

}

func TestOrderBy(t *testing.T) {
	tests := []struct {
		name   string
		sql    string
		want   string
		params map[string]interface{}
	}{
		{
			name: "order by column, default",
			sql:  "select column1 from table order by column1",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"column1\" ASC",
		},
		{
			name: "order by c1,c2 desc,",
			sql:  "select column1 from table order by column2 desc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"number_field_values[indexOf(number_field_keys,'column2')]\" DESC",
		},
		{
			name: "order by column desc, default",
			sql:  "select column1 from table order by column1 desc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"column1\" DESC",
		},
		{
			name: "order by column desc, column2 asc",
			sql:  "select column1,column2 from table order by column1 desc,column2 asc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\", toNullable(number_field_values[indexOf(number_field_keys,'column2')]) AS \"column2\" FROM \"table\" ORDER BY \"column1\" DESC,\"column2\" ASC",
		},
		{
			name: "order by column1 asc, column2 desc",
			sql:  "select column1,column2 from table order by column1 asc,column2 desc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\", toNullable(number_field_values[indexOf(number_field_keys,'column2')]) AS \"column2\" FROM \"table\" ORDER BY \"column1\" ASC,\"column2\" DESC",
		},
		{
			name: "none order by",
			sql:  "select column1 from table group by column1",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"column1\" ASC",
		},
		{
			name: "asc",
			sql:  "select column1 from table order by column1 asc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"column1\" ASC",
		},
		{
			name: "desc",
			sql:  "select column1 from table order by column1 desc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"column1\" DESC",
		},
		{
			name: "max",
			sql:  "select service_id::tag,max(timestamp) from table GROUP BY service_id::tag ORDER BY max(timestamp) DESC",
			want: "SELECT MAX(timestamp) AS \"1362043e612fc3f5\", toNullable(tag_values[indexOf(tag_keys,'service_id')]) AS \"service_id::tag\" FROM \"table\" ORDER BY \"1362043e612fc3f5\" DESC,\"service_id::tag\" ASC",
		},
		{
			name: "no group, no y",
			sql:  "select avg(load1::field),avg(load5::field),avg(load15::field) from table GROUP BY time()",
			want: `SELECT AVG(if(indexOf(number_field_keys,'load5') == 0,null,number_field_values[indexOf(number_field_keys,'load5')])) AS "2756505adaab8174", AVG(if(indexOf(number_field_keys,'load15') == 0,null,number_field_values[indexOf(number_field_keys,'load15')])) AS "f401d9b077f06a6e", AVG(if(indexOf(number_field_keys,'load1') == 0,null,number_field_values[indexOf(number_field_keys,'load1')])) AS "719d37bba7262d89", MIN("timestamp") AS "bucket_timestamp" FROM "table" GROUP BY intDiv(toRelativeSecondNum(timestamp), 60) ORDER BY "bucket_timestamp" ASC`,
		},
		{
			name: "group. axis",
			sql:  "select avg(load1::field),avg(load5::field),avg(load15::field),service_id::tag from table GROUP BY service_id::tag,time()",
			want: `SELECT AVG(if(indexOf(number_field_keys,'load15') == 0,null,number_field_values[indexOf(number_field_keys,'load15')])) AS "f401d9b077f06a6e", AVG(if(indexOf(number_field_keys,'load1') == 0,null,number_field_values[indexOf(number_field_keys,'load1')])) AS "719d37bba7262d89", AVG(if(indexOf(number_field_keys,'load5') == 0,null,number_field_values[indexOf(number_field_keys,'load5')])) AS "2756505adaab8174", toNullable(tag_values[indexOf(tag_keys,'service_id')]) AS "service_id::tag", MIN("timestamp") AS "bucket_timestamp" FROM "table" GROUP BY "service_id::tag", intDiv(toRelativeSecondNum(timestamp), 60) ORDER BY "service_id::tag" ASC, "bucket_timestamp" ASC`,
		},
		{
			name: "avg column",
			sql:  "select avg(cpu_usage_percent::field) from table GROUP BY time()",
			want: "SELECT AVG(if(indexOf(number_field_keys,'cpu_usage_percent') == 0,null,number_field_values[indexOf(number_field_keys,'cpu_usage_percent')])) AS \"8c8e2100c8bcb2e3\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"8c8e2100c8bcb2e3\", intDiv(toRelativeSecondNum(timestamp), 60) ORDER BY \"bucket_timestamp\" ASC",
		},
		{
			name: "timestamp should by last order",
			sql:  "select time(),service_id::tag from table GROUP BY service_id::tag,time() ORDER BY service_id::tag desc",
			want: "SELECT toNullable(tag_values[indexOf(tag_keys,'service_id')]) AS \"service_id::tag\" FROM \"table\" ORDER BY \"service_id::tag\" DESC,\"bucket_timestamp\" ASC",
		},
		{
			name: "timestamp desc",
			sql:  "select column1 from table order by timestamp desc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"timestamp\" DESC",
		},
		{
			name: "timestamp asc",
			sql:  "select column1 from table order by timestamp asc",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column1')]) AS \"column1\" FROM \"table\" ORDER BY \"timestamp\" ASC",
		},
		{
			name: "test",
			sql:  "SELECT key::tag,max(value) as max_value FROM jvm_gc GROUP BY time(),key::tag",
			want: "SELECT MAX(value) AS \"max_value\", key AS \"key::tag\", addSeconds(toDateTime('1970-01-01 00:00:00'),intDiv(toRelativeSecondNum(timestamp), 60) * 60) AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"key::tag\", intDiv(toRelativeSecondNum(timestamp), 60) ORDER BY \"key::tag\" ASC, \"bucket_timestamp\" ASC",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Parser{
				ctx: &Context{
					dimensions: make(map[string]bool),
					timeKey:    "timestamp",
				},
			}
			parse := influxql.NewParser(strings.NewReader(test.sql))
			if len(test.params) > 0 {
				parse.SetParams(test.params)
			}
			q, err := parse.ParseQuery()

			require.NoError(t, err)
			selectStmt, ok := q.Statements[0].(*influxql.SelectStatement)

			expr := goqu.From("table")
			require.Truef(t, ok, "parse query is not select statement")

			expr, handlers, columns, err := p.parseQueryOnExpr(selectStmt.Fields, expr)

			expr, groupColumns, err := p.ParseGroupByOnExpr(selectStmt.Dimensions, expr, &handlers, columns)
			require.NoError(t, err, "parse group by error")

			expr, err = p.ParseOrderByOnExpr(selectStmt.SortFields, expr, columns, groupColumns)

			require.NoError(t, err)

			sql, _, err := expr.ToSQL()
			require.NoError(t, err)

			t.Log(test.sql)
			t.Log("got", sql)
			t.Log("want", test.want)

			if strings.Index(test.want, "ORDER BY") != -1 {
				// column is map, maybe disorder by order
				orderWantExpr := strings.ReplaceAll(test.want[strings.Index(test.want, "ORDER BY")+8:], " ", "")
				orderGotExpr := strings.ReplaceAll(sql[strings.Index(sql, "ORDER BY")+8:], " ", "")

				got := strings.Split(orderGotExpr, ",")
				want := strings.Split(orderWantExpr, ",")
				require.Equalf(t, len(want), len(got), "length should by equal, should by %v, but is %v", orderWantExpr, orderGotExpr)

				for i, w := range want {
					require.Equalf(t, w, got[i], "[%v] not equal", i)
				}

			}
		})
	}
}

func TestGroupColumnShouldBeExist(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "normal single group",
			sql:  "select column from table group by column",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column\" FROM \"table\" WHERE has(tag_keys,'column') GROUP BY \"column\"",
		},
		{
			name: "normal single group tag column",
			sql:  "select column::tag from table group by column::tag",
			want: "SELECT toNullable(tag_values[indexOf(tag_keys,'column')]) AS \"column::tag\" FROM \"table\" WHERE has(tag_keys,'column') GROUP BY \"column::tag\"",
		},
		{
			name: "normal single group field column",
			sql:  "select column::field from table group by column::field",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column::field\" FROM \"table\" WHERE has(tag_keys,'column') GROUP BY \"column::field\"",
		},
		{
			name: "column and agg function",
			sql:  "select column::field,max(column) from table group by column::field",
			want: "SELECT MAX(number_field_values[indexOf(number_field_keys,'column')]) AS \"322cc30ad1d92b84\", toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column::field\" FROM \"table\" WHERE has(tag_keys,'column') GROUP BY \"column::field\"",
		},
		{
			name: "time",
			sql:  "select column::field from table group by time()",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column::field\", MIN(\"timestamp\") AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"column::field\", intDiv(toRelativeSecondNum(timestamp), 60)",
		},
		{
			name: "origin column",
			sql:  "select tenant_id,timestamp,org_name from table group by tenant_id,timestamp,org_name",
			want: "SELECT toNullable(tenant_id) AS \"tenant_id\", toNullable(timestamp) AS \"timestamp\", toNullable(org_name) AS \"org_name\" FROM \"table\" GROUP BY \"tenant_id\", \"timestamp\", \"org_name\"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Parser{
				ctx: &Context{
					originalTimeUnit: tsql.Nanosecond,
					targetTimeUnit:   tsql.Millisecond,
					dimensions:       make(map[string]bool),
				},
			}
			parse := influxql.NewParser(strings.NewReader(test.sql))

			q, err := parse.ParseQuery()

			require.NoError(t, err)
			selectStmt, ok := q.Statements[0].(*influxql.SelectStatement)

			expr := goqu.From("table")
			require.Truef(t, ok, "parse query is not select statement")

			expr, handler, columns, err := p.parseQueryOnExpr(selectStmt.Fields, expr)

			expr, _, err = p.ParseGroupByOnExpr(selectStmt.Dimensions, expr, &handler, columns)

			require.NoError(t, err)

			sql, _, err := expr.ToSQL()
			require.NoError(t, err)

			t.Log(test.sql)
			t.Log("got", sql)
			t.Log("want", test.want)
			sql = sql[strings.Index(sql, "GROUP BY")+8:]
			test.want = test.want[strings.Index(test.want, "GROUP BY")+8:]

			require.ElementsMatch(t, strings.Split(test.want, ","), strings.Split(sql, ","))

		})
	}

}
