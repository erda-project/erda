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

func (f *Formater) formatTableChart(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	headers := make([]map[string]interface{}, len(rs.Columns), len(rs.Columns))
	for i, c := range rs.Columns {
		headers[i] = map[string]interface{}{
			"title":     c.Name,
			"dataIndex": strconv.Itoa(i),
		}
	}
	list := make([]map[string]interface{}, 0)
	for _, row := range rs.Rows {
		data := make(map[string]interface{}, len(row))
		for i, v := range row {
			data[strconv.Itoa(i)] = v
		}
		list = append(list, data)
	}
	return map[string]interface{}{
		"metricData": list,
		"cols":       headers,
	}, nil
}

func (f *Formater) formatTableChartV2(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	headers := make([]map[string]interface{}, len(rs.Columns), len(rs.Columns))
	for i, c := range rs.Columns {
		col := map[string]interface{}{
			"key":  c.Name,
			"flag": c.Flag.String(),
		}
		if c.Key != c.Name {
			col["_key"] = c.Key
		}
		headers[i] = col
	}
	list := make([]map[string]interface{}, 0)
	for _, row := range rs.Rows {
		data := make(map[string]interface{}, len(row))
		for i, v := range row {
			col := rs.Columns[i]
			data[col.Name] = v
		}
		list = append(list, data)
	}
	return map[string]interface{}{
		"data":     list,
		"cols":     headers,
		"interval": rs.Interval,
	}, nil
}
