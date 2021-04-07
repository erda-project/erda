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

import tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"

func (f *Formater) formatCardChart(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	var list []map[string]interface{}
	if len(rs.Columns) <= 0 {
		return list, nil
	}
	for _, row := range rs.Rows {
		data := map[string]interface{}{
			"name":  rs.Columns[0].Name,
			"value": row[0],
		}
		list = append(list, data)
	}
	return list, nil
}
