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
			want: "SELECT * FROM \"table\" WHERE ((column = '12313') OR (column = '12313'))",
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
			want: "SELECT * FROM \"table\" WHERE ((column IN ('111', '222', '3333')) OR (column IN ('111', '222', '3333')))",
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
			want: "SELECT * FROM \"table\" WHERE (((column = '12313') AND (column = '12313')) OR ((column = '12313') OR (column = '12313')))",
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

			expr, err := p.filterToExpr(test.args, expr)
			require.NoError(t, err)
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
			name: "select column",
			sql:  "select column from table",
			want: "SELECT number_field_values[indexOf(number_field_keys,'column')] AS \"column\" FROM \"table\"",
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
			want: "SELECT count(distinct(tag_values[indexOf(tag_keys,'service_id')])) AS \"2e1cf76dc42cffe8\" FROM \"table\"",
		},
		{
			name: "select max",
			sql:  "select max(column) from table",
			want: "SELECT MAX(number_field_values[indexOf(number_field_keys,'column')]) AS \"322cc30ad1d92b84\" FROM \"table\"",
		},
		{
			name: "select min",
			sql:  "select min(service_id::tag) from table",
			want: "SELECT MIN(tag_values[indexOf(tag_keys,'service_id')]) AS \"b784dcb694669c75\" FROM \"table\"",
		},
		{
			name: "select count",
			sql:  "select count(service_id::tag) from table",
			want: "SELECT COUNT(if(indexOf(tag_keys,'service_id') = 1,1,null)) AS \"993567cffecb105e\" FROM \"table\"",
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
			want: "SELECT AVG(number_field_values[indexOf(number_field_keys,'committed')]) AS \"cd848468318e898b\" FROM \"table\"",
		},
	}

	for _, test := range tests {
		//parseQueryFieldsByExpr
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
			want: "SELECT MAX(number_field_values[indexOf(number_field_keys,'column')]) AS \"322cc30ad1d92b84\", toDateTime64(toStartOfInterval(timestamp, toIntervalSecond(60),'UTC'),9) AS \"bucket_timestamp\" FROM \"table\" GROUP BY bucket_timestamp",
		},
		{
			name: "time(2h)",
			sql:  "select column from table group by time(2h)",
			want: "SELECT number_field_values[indexOf(number_field_keys,'column')] AS \"column\", toDateTime64(toStartOfInterval(timestamp, toIntervalSecond(7200),'UTC'),9) AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"column\", bucket_timestamp",
		},
		{
			name: "groupby,time()",
			sql:  "select sum(http_status_code_count::field),http_status_code::tag from table GROUP BY time(),http_status_code::tag",
			want: "SELECT SUM(number_field_values[indexOf(number_field_keys,'http_status_code_count')]) AS \"339c9df3d700c4f0\", tag_values[indexOf(tag_keys,'http_status_code')] AS \"http_status_code\", toDateTime64(toStartOfInterval(timestamp, toIntervalSecond(60),'UTC'),9) AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"http_status_code\", bucket_timestamp",
		},
		{
			name: "time(),column",
			sql:  "select http_status_code::tag from table GROUP BY time()",
			want: "SELECT tag_values[indexOf(tag_keys,'http_status_code')] AS \"http_status_code\", toDateTime64(toStartOfInterval(timestamp, toIntervalSecond(60),'UTC'),9) AS \"bucket_timestamp\" FROM \"table\" GROUP BY \"http_status_code\", bucket_timestamp",
		},
		{
			name: "no group",
			sql:  "select max(http_status_code::tag) from table",
			want: "SELECT MAX(tag_values[indexOf(tag_keys,'http_status_code')]) AS \"61421335fd474c8e\" FROM \"table\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := Parser{
				ctx: &Context{
					originalTimeUnit: tsql.Nanosecond,
					targetTimeUnit:   tsql.Millisecond,
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

			t.Log(sql)
			if strings.Index(test.want, strings.ToUpper("group by")) <= 0 {
				require.Equal(t, test.want, sql)
				return
			}
			sqlGroup := strings.Split(sql, strings.ToUpper("group by"))[1]
			wantGroup := strings.Split(test.want, strings.ToUpper("group by"))[1]

			require.ElementsMatch(t, strings.Split(wantGroup, ","), strings.Split(sqlGroup, ","))
			//require.Equal(t, test.want, sql)
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
			want: "SELECT number_field_values[indexOf(number_field_keys,'column1')] AS \"column1\" FROM \"table\" ORDER BY \"column1\" ASC",
		},
		{
			name: "none order by",
			sql:  "select column1 from table",
			want: "SELECT number_field_values[indexOf(number_field_keys,'column1')] AS \"column1\" FROM \"table\"",
		},
		{
			name: "asc",
			sql:  "select column1 from table order by column1 asc",
			want: "SELECT number_field_values[indexOf(number_field_keys,'column1')] AS \"column1\" FROM \"table\" ORDER BY \"column1\" ASC",
		},
		{
			name: "desc",
			sql:  "select column1 from table order by column1 desc",
			want: "SELECT number_field_values[indexOf(number_field_keys,'column1')] AS \"column1\" FROM \"table\" ORDER BY \"column1\" DESC",
		},
		{
			name: "max",
			sql:  "select service_id::tag,max(timestamp) from table GROUP BY service_id::tag ORDER BY max(timestamp) DESC",
			want: "SELECT MAX(timestamp) AS \"1362043e612fc3f5\", tag_values[indexOf(tag_keys,'service_id')] AS \"service_id\" FROM \"table\" ORDER BY \"1362043e612fc3f5\" DESC",
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

			expr, _, columns, err := p.parseQueryOnExpr(selectStmt.Fields, expr)

			expr, err = p.ParseOrderByOnExpr(selectStmt.SortFields, expr, columns)

			require.NoError(t, err)

			sql, _, err := expr.ToSQL()

			require.NoError(t, err)
			require.Equal(t, test.want, sql)
		})
	}
}
