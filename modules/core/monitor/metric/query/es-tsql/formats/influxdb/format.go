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
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	tsql "github.com/erda-project/erda/modules/core/monitor/metric/query/es-tsql"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/es-tsql/formats"
	"github.com/erda-project/erda/pkg/common/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

// Formater .
type Formater struct{}

// Response .
type Response struct {
	Results []*pb.Result `json:"results,omitempty"`
	Error   error        `json:"error,omitempty"`
}

// Format .
func (f *Formater) Format(q tsql.Query, rs *tsql.ResultSet, params map[string]interface{}) (interface{}, error) {
	var columns []string
	for _, c := range rs.Columns {
		columns = append(columns, c.Name)
	}
	rows := make([]*pb.Row, len(rs.Rows))
	for i, values := range rs.Rows {
		vals := make([]*structpb.Value, len(values))
		for i, val := range values {
			if val != nil {
				val, err := structpb.NewValue(val)
				if err != nil {
					return nil, errors.NewInternalServerError(err)
				}
				vals[i] = val
			}
		}
		rows[i] = &pb.Row{Values: vals}
	}
	return api.SuccessRaw(&Response{
		Results: []*pb.Result{
			{
				StatementId: 0,
				Series: []*pb.Serie{
					{
						Name:    getSourceName(q),
						Columns: columns,
						Rows:    rows,
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
