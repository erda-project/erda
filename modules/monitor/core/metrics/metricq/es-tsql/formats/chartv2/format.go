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
	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql/formats"
)

// Formater .
type Formater struct{}

// Format .
func (f *Formater) Format(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	typ := "table"
	if t, ok := params["type"].(string); ok {
		typ = t
	}
	switch typ {
	case "pie", "list":
		return f.formatListChart(q, rs, params)
	case "line":
		return f.formatLineChart(q, rs, params)
	case "bar":
		return f.formatBarChart(q, rs, params)
	case "card":
		return f.formatCardChart(q, rs, params)
	case "*", "_":
		return f.formatTableChartV2(q, rs, params)
	}
	return f.formatTableChart(q, rs, params)
}

var formater Formater

func init() {
	formats.RegisterFormater("chartv2", &formater)
}
