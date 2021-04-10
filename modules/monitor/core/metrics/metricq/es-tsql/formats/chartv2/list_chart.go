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

func (f *Formater) formatListChart(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	list := make([]map[string]interface{}, 0)
	var title map[string]bool
	if val, ok := params["title"].(string); ok && len(val) > 0 {
		title = make(map[string]bool)
		for _, v := range strings.Split(val, ",") {
			v = strings.TrimSpace(v)
			if len(v) > 0 {
				title[v] = true
			}
		}
	}
	if len(title) > 0 {
		for _, row := range rs.Rows {
			data := make(map[string]interface{}, 2)
			var headers []string
			var values []interface{}
			for i, c := range rs.Columns {
				if title[c.Name] {
					if row[i] == nil {
						headers = append(headers, "")
					} else {
						headers = append(headers, fmt.Sprint(row[i]))
					}
				} else {
					values = append(values, row[i])
				}
			}
			data["title"] = strings.Join(headers, ",")
			if len(values) == 0 {
				data["value"] = nil
			} else if len(values) == 1 {
				data["value"] = values[0]
			} else {
				data["value"] = values
			}
			list = append(list, data)
		}
	} else {
		if len(rs.Columns) == 1 {
			for i, row := range rs.Rows {
				data := make(map[string]interface{}, 2)
				data["title"] = strconv.Itoa(i + 1)
				data["value"] = row[0]
				list = append(list, data)
			}
		} else {
			for i, row := range rs.Rows {
				data := make(map[string]interface{}, 2)
				data["title"] = strconv.Itoa(i + 1)
				data["value"] = row
				list = append(list, data)
			}
		}
	}
	return map[string]interface{}{
		"metricData": list,
	}, nil
}
