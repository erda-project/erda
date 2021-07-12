// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package metric

import (
	"strconv"
	"testing"
)

func Test_getMetricFromSQL(t *testing.T) {
	tests := []struct {
		sql  string
		want string
	}{
		{
			sql:  "SELECT * FROM metric1",
			want: "metric1",
		},
		{
			sql:  "select * FrOM metric1",
			want: "metric1",
		},
		{
			sql:  "sElEct * from metric1",
			want: "metric1",
		},
		{
			sql:  "SELECT field from metric1",
			want: "metric1",
		},
		{
			sql:  "SELECT *from metric1",
			want: "",
		},
		{
			sql:  "SELECT * from metric1;",
			want: "metric1",
		},
		{
			sql:  "SELECT field from metric1 ;",
			want: "metric1",
		},
		{
			sql:  "SELECT * from metric1 WHERE id::tag='1';",
			want: "metric1",
		},
		{
			sql:  "SELECT * from metric1WHERE id::tag='1';",
			want: "metric1WHERE",
		},
		{
			sql:  "SELECT field from metric1,metric2",
			want: "metric1",
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := getMetricFromSQL(tt.sql); got != tt.want {
				t.Errorf("getMetricFromSQL(%q) = %v, want %v", tt.sql, got, tt.want)
			}
		})
	}
}
