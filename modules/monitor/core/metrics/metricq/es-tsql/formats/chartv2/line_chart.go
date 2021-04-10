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

package chartv2

import (
	"fmt"
	"strconv"
	"strings"

	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
)

type aggData struct {
	Name string        `json:"name"`
	Tag  interface{}   `json:"tag"`
	Data []interface{} `json:"data"`
}

func (f *Formater) formatLineChart(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	timeIdx, _, dims, vals := getGroupColumns(rs.Columns)
	var list []interface{}
	var times []int64
	groups := make(map[string]map[string]*aggData)
	var data map[string]*aggData
	for _, row := range rs.Rows {
		var tag interface{}
		if len(dims) > 0 {
			key := f.getGroupKey(dims, row)
			group, ok := groups[key]
			if !ok {
				group = make(map[string]*aggData)
				groups[key] = group
				list = append(list, group)
			}
			data, tag = group, key
		} else if data == nil {
			data = make(map[string]*aggData)
			list = append(list, data)
		}
		if timeIdx >= 0 && len(groups) <= 1 {
			ts, _ := tsql.GetTimestampValue(row[timeIdx])
			ts = tsql.ConvertTimestamp(ts, q.Context().TargetTimeUnit(), tsql.Millisecond) // 强制转换成 ms
			times = append(times, ts)
		}
		for _, val := range vals {
			key := strconv.Itoa(val)
			agg, ok := data[key]
			if !ok {
				col := rs.Columns[val]
				agg = &aggData{
					Name: col.Name,
					Tag:  tag,
				}
				data[key] = agg
			}
			agg.Data = append(agg.Data, row[val])
		}
	}
	return map[string]interface{}{
		"time": times,
		"results": []interface{}{
			map[string]interface{}{
				"data": list,
			},
		},
	}, nil
}

func (f *Formater) getGroupKey(idxs []int, row []interface{}) string {
	var group []string
	for _, i := range idxs {
		v := row[i]
		if v == nil {
			group = append(group, "")
		} else {
			group = append(group, fmt.Sprint(v))
		}
	}
	return strings.Join(group, ",")
}

func getGroupColumns(cols []*tsql.Column) (ti, rngs int, dims, vals []int) {
	rngs, ti = -1, -1
	for i, c := range cols {
		if c.Flag&tsql.ColumnFlagGroupBy == tsql.ColumnFlagGroupBy {
			if c.Flag&tsql.ColumnFlagGroupByInterval == tsql.ColumnFlagGroupByInterval {
				ti = i
			} else if c.Flag&tsql.ColumnFlagGroupByRange == tsql.ColumnFlagGroupByRange {
				rngs = i
			} else {
				dims = append(dims, i)
			}
		} else {
			vals = append(vals, i)
		}
	}
	return
}
