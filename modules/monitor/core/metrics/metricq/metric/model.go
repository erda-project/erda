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

package metric

import "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"

type InfluxqlRespone struct {
	Result []*Results `json:"results"`
}

type Results struct {
	StatementId int       `json:"statement_id"`
	Series      []*Series `json:"series"`
}

type Series struct {
	Name    string          `json:"name"`
	Columns []string        `json:"columns"`
	Values  [][]interface{} `json:"values"`
}

type TableResponse struct {
	Cols     []*pb.TableColumn        `json:"cols"`
	Data     []map[string]interface{} `json:"data"`
	Interval int64                    `json:"interval"`
}
