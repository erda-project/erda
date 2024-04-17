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
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/influxdata/influxql"
	"github.com/stretchr/testify/require"

	tsql "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/es-tsql"
)

func TestAggregationFunction(t *testing.T) {
	tests := []struct {
		name     string
		function string
		stmt     influxql.Call
		want     string
	}{
		{
			name:     "test diffps",
			function: "diffps",
			stmt: influxql.Call{
				Name: "diffps",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT Min(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"0e4b5b1082a11703\" FROM \"table\"",
		},
		{
			name:     "test diff",
			function: "diff",
			stmt: influxql.Call{
				Name: "diffps",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT Min(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"0e4b5b1082a11703\" FROM \"table\"",
		},
		{
			name:     "test max",
			function: "max",
			stmt: influxql.Call{
				Name: "max",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT MAX(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"1af26e55a4a29af8\" FROM \"table\"",
		},
		{
			name:     "test min",
			function: "min",
			stmt: influxql.Call{
				Name: "min",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT MIN(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"92b276ffba2e062b\" FROM \"table\"",
		},
		{
			name:     "test avg",
			function: "avg",
			stmt: influxql.Call{
				Name: "diffps",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT AVG(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"0e4b5b1082a11703\" FROM \"table\"",
		},
		{
			name:     "test count",
			function: "count",
			stmt: influxql.Call{
				Name: "count",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT COUNT(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"e1910da5a36c4a26\" FROM \"table\"",
		},
		{
			name:     "test distinct",
			function: "distinct",
			stmt: influxql.Call{
				Name: "distinct",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT uniqCombined(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"f316cf2c588e6404\" FROM \"table\"",
		},
		{
			name:     "test rateps",
			function: "rateps",
			stmt: influxql.Call{
				Name: "distinct",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT SUM(number_field_values[indexOf(number_field_keys,'com_delete')]) AS \"f316cf2c588e6404\" FROM \"table\"",
		},
		{
			name:     "test value",
			function: "value",
			stmt: influxql.Call{
				Name: "value",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT MAX(number_field_values[indexOf(number_field_keys,'com_delete')]) AS \"f18a4b3b5bedb176\" FROM \"table\"",
		},
		{
			name:     "test first",
			function: "first",
			stmt: influxql.Call{
				Name: "distinct",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT argMin(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')]),timestamp) AS \"f316cf2c588e6404\" FROM \"table\"",
		},
		{
			name:     "test last",
			function: "last",
			stmt: influxql.Call{
				Name: "distinct",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			want: "SELECT anyLast(if(indexOf(number_field_keys,'com_delete') == 0,null,number_field_values[indexOf(number_field_keys,'com_delete')])) AS \"f316cf2c588e6404\" FROM \"table\"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, ok := CkAggFunctions[test.function]
			if !ok {
				t.Errorf("should by implemented diffps")
			}
			ctx := Context{}
			ctx.timeKey = "timestamp"
			id := ctx.GetFuncID(&test.stmt, influxql.AnyField)
			define, err := p.New(&ctx, id, &test.stmt)
			require.NoError(t, err)
			parser := Parser{}

			aggs := make(map[string]exp.Expression)
			err = define.Aggregations(&parser, aggs, FuncFlagSelect)

			require.Equal(t, 1, len(aggs))

			v, ok := aggs[id]
			require.True(t, ok)
			expr := goqu.From("table")
			expr = expr.Select(v)
			sql, _, err := expr.ToSQL()
			require.NoError(t, err)
			require.Equal(t, test.want, sql)

		})
	}
}

func TestAggregationHandlerFunction(t *testing.T) {
	tests := []struct {
		name       string
		function   string
		id         string
		stmt       influxql.Call
		nextRow    map[string]interface{}
		want       interface{}
		currentRow map[string]interface{}
	}{
		{
			name:     "test diffps, no next",
			function: "diffps",
			id:       "0e4b5b1082a11703",
			stmt: influxql.Call{
				Name: "diffps",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			currentRow: map[string]interface{}{
				"0e4b5b1082a11703": float64(111),
			},
			want: 0,
		},
		{
			name:     "test diffps, nil",
			function: "diffps",
			id:       "0e4b5b1082a11703",
			stmt: influxql.Call{
				Name: "diffps",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			currentRow: map[string]interface{}{
				"0e4b5b1082a11703": nil,
			},
			want: 0,
		},
		{
			name:     "test diffps",
			function: "diffps",
			id:       "0e4b5b1082a11703",
			stmt: influxql.Call{
				Name: "diffps",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			currentRow: map[string]interface{}{
				"0e4b5b1082a11703": float64(222),
			},
			nextRow: map[string]interface{}{
				"0e4b5b1082a11703": float64(111),
			},
			want: 0,
		},

		{
			name:     "test diff, no next",
			function: "diff",
			id:       "0e4b5b1082a11703",
			stmt: influxql.Call{
				Name: "diff",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			currentRow: map[string]interface{}{
				"0e4b5b1082a11703": float64(111),
			},
			want: 0,
		},
		{
			name:     "test diff, nil",
			function: "diff",
			id:       "0e4b5b1082a11703",
			stmt: influxql.Call{
				Name: "diff",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			currentRow: map[string]interface{}{
				"0e4b5b1082a11703": nil,
			},
			want: 0,
		},
		{
			name:     "test diff",
			function: "diff",
			id:       "0e4b5b1082a11703",
			stmt: influxql.Call{
				Name: "diff",
				Args: []influxql.Expr{
					&influxql.VarRef{
						Val:  "com_delete",
						Type: influxql.AnyField,
					},
				},
			},
			currentRow: map[string]interface{}{
				"0e4b5b1082a11703": float64(222),
			},
			nextRow: map[string]interface{}{
				"0e4b5b1082a11703": float64(111),
			},
			want: 111.0 - 222.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p, ok := CkAggFunctions[test.function]
			if !ok {
				t.Errorf("should by implemented diffps")
			}
			ctx := Context{}
			ctx.interval = 60
			ctx.targetTimeUnit = tsql.Second
			ctx.AttributesCache()
			ctx.attributesCache["next"] = test.nextRow

			define, err := p.New(&ctx, test.id, &test.stmt)
			require.NoError(t, err)

			result, err := define.Handle(test.currentRow)

			require.NoError(t, err)
			require.Equal(t, test.want, result)

		})
	}
}
