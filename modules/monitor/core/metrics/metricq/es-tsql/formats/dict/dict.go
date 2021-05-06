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

package dict

import (
	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql/formats"
)

// Formater .
type Formater struct{}

// Format .
func (f *Formater) Format(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	res := make([]map[string]interface{}, len(rs.Rows))

	for i := 0; i < len(rs.Rows); i++ {
		for j := 0; j < len(rs.Columns); j++ {
			key, val := rs.Columns[j].Name, rs.Rows[i][j]
			if res[i] == nil {
				res[i] = map[string]interface{}{
					key: val,
				}
			} else {
				res[i][key] = val
			}
		}
	}
	return res, nil
}

var formater Formater

func init() {
	formats.RegisterFormater("dict", &formater)
}
