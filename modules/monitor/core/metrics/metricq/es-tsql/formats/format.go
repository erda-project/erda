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

package formats

import (
	"fmt"

	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
)

// Formater response formater
type Formater interface {
	Format(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error)
}

// Formats .
var Formats = map[string]Formater{}

// RegisterFormater .
func RegisterFormater(name string, formater Formater) {
	Formats[name] = formater
}

// Format .
func Format(name string, q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	f, ok := Formats[name]
	if !ok {
		return nil, fmt.Errorf("invalid formater '%s'", name)
	}
	return f.Format(q, rs, params)
}
