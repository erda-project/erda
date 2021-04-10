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
	"strconv"

	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
)

func (f *Formater) formatBarChart(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	timeIdx, _, dims, vals := getGroupColumns(rs.Columns)
	if timeIdx >= 0 {
		dims = append([]int{timeIdx}, dims...)
	}
	var list []interface{}
	var xdata []interface{}
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
			xdata = append(xdata, key)
		} else if data == nil {
			data = make(map[string]*aggData)
			list = append(list, data)
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
		"xData": xdata,
		"results": []interface{}{
			map[string]interface{}{
				"data": list,
			},
		},
	}, nil
}
