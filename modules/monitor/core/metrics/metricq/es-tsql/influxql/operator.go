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

package esinfluxql

import (
	tsql "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql"
	"github.com/influxdata/influxql"
)

func reverseOperator(op influxql.Token) influxql.Token {
	switch op {
	case influxql.LT:
		return influxql.GT
	case influxql.LTE:
		return influxql.GTE
	case influxql.GT:
		return influxql.LT
	case influxql.GTE:
		return influxql.LTE
	}
	return op
}

func toOperator(op influxql.Token) tsql.Operator {
	return tsql.Operator(int(op) - int(influxql.ADD) + 1)
}
