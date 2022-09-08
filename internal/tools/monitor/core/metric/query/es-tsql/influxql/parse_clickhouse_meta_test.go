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
	"strings"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/influxdata/influxql"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
)

type mockMeta struct {
	mockMeta []*pb.MetricMeta
}

func (m mockMeta) GetMetricMetaByCache(scope, scopeID string, names ...string) ([]*pb.MetricMeta, error) {
	return m.mockMeta, nil
}

func TestMeta(t *testing.T) {
	tests := []struct {
		name     string
		mockMeta []*pb.MetricMeta
		sql      string
		want     string
	}{
		{
			name: "number field",
			sql:  "select http_status_code_count::field from metric",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'http_status_code_count')]) AS \"http_status_code_count::field\" FROM \"metric\" ",
		},
		{
			name: "string field",
			mockMeta: []*pb.MetricMeta{
				{
					Fields: map[string]*pb.FieldDefine{
						"http_status_code_count": {
							Type: "string",
						},
					},
				},
			},
			sql:  "select http_status_code_count::field from metric",
			want: "SELECT toNullable(string_field_values[indexOf(string_field_keys,'http_status_code_count')]) AS \"http_status_code_count::field\" FROM \"metric\" ",
		},
		{
			name: "select string function field",
			mockMeta: []*pb.MetricMeta{
				{
					Fields: map[string]*pb.FieldDefine{
						"http_status_code_count": {
							Type: "string",
						},
					},
				},
			},
			sql:  "select if(gt(http_status_code_count::field-timestamp,300000000000),'false','true') from metric",
			want: "SELECT toNullable(string_field_values[indexOf(string_field_keys,'http_status_code_count')]) AS \"http_status_code_count::field\", toNullable(timestamp) AS \"timestamp\" FROM \"metric\" ",
		},
		{
			name: "where string function field",
			mockMeta: []*pb.MetricMeta{
				{
					Fields: map[string]*pb.FieldDefine{
						"http_status_code_count": {
							Type: "string",
						},
					},
				},
			},
			sql:  "select column from metric where http_status_code_count::field > 100",
			want: "SELECT toNullable(number_field_values[indexOf(number_field_keys,'column')]) AS \"column\" FROM \"metric\" WHERE ((string_field_values[indexOf(string_field_keys,'http_status_code_count')] > 100))",
		},
		{
			name: "no column name",
			mockMeta: []*pb.MetricMeta{
				{
					Fields: map[string]*pb.FieldDefine{
						"column": {
							Type: "string",
						},
					},
				},
			},
			sql:  "select column from metric",
			want: "SELECT toNullable(string_field_values[indexOf(string_field_keys,'column')]) AS \"column\" FROM \"metric\" ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parse := Parser{
				ctx: &Context{
					timeKey: "timestamp",
				},
			}
			parse.SetMeta(mockMeta{
				mockMeta: test.mockMeta,
			})

			stm, err := influxql.NewParser(strings.NewReader(test.sql)).ParseQuery()

			require.NoError(t, err)
			selectStmt, ok := stm.Statements[0].(*influxql.SelectStatement)
			require.True(t, ok)

			plan, err := parse.ParseClickhouse(selectStmt)
			require.NoError(t, err)

			execPlan, ok := plan.SearchSource().(*goqu.SelectDataset)
			require.True(t, ok)

			sql, _, err := execPlan.ToSQL()
			require.NoError(t, err)

			sql = strings.ReplaceAll(sql, "(\"timestamp\" >= fromUnixTimestamp64Nano(cast(0,'Int64')))", "")
			sql = strings.ReplaceAll(sql, "(\"timestamp\" < fromUnixTimestamp64Nano(cast(0,'Int64')))", "")
			sql = strings.ReplaceAll(sql, "( AND )", "")
			sql = strings.ReplaceAll(sql, " AND  AND ", "")
			sql = strings.TrimSpace(sql)

			if strings.HasSuffix(sql, "WHERE") {
				sql = strings.ReplaceAll(sql, "WHERE", "")
			}

			require.Equal(t, test.want, sql)

		})
	}
}
