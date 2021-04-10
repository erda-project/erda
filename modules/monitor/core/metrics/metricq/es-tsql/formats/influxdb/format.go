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

package influxdb

import (
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql/formats"
)

// Formater .
type Formater struct{}

// Response .
type Response struct {
	Results []interface{} `json:"results,omitempty"`
	Error   error         `json:"error,omitempty"`
}

// Format .
func (f *Formater) Format(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	var columns []string
	for _, c := range rs.Columns {
		columns = append(columns, c.Name)
	}
	return api.SuccessRaw(&Response{
		Results: []interface{}{
			map[string]interface{}{
				"statement_id": 0,
				"series": []interface{}{
					map[string]interface{}{
						"name":    getSourceName(q),
						"columns": columns,
						"values":  rs.Rows,
					},
				},
			},
		},
	}), nil
}

func getSourceName(q tsql.Query) string {
	var list []string
	for _, s := range q.Sources() {
		if len(s.Name) > 0 {
			list = append(list, s.Name)
		}
	}
	return strings.Join(list, ",")
}

var formater Formater

func init() {
	formats.RegisterFormater("influxdb", &formater)
}
