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

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	tsql "github.com/erda-project/erda/modules/core/monitor/metric/query/es-tsql"
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
	if _, ok := params["protobuf"]; !ok {
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
	headers := make([]*pb.TableColumn, len(rs.Columns), len(rs.Columns))
	for i, c := range rs.Columns {
		col := &pb.TableColumn{
			Flag: c.Flag.String(),
			Key:  c.Name, // TODO: change to c.Key
			Name: c.Key,  // TODO: change to c.Name
		}
		headers[i] = col
	}
	list := make([]*pb.TableRow, 0)
	for _, row := range rs.Rows {
		data := &pb.TableRow{
			Values: make(map[string]*structpb.Value),
		}
		for i, v := range row {
			col := rs.Columns[i]
			if v != nil {
				val, err := structpb.NewValue(v)
				if err != nil {
					return nil, fmt.Errorf("convert value: %w", err)
				}
				data.Values[col.Name] = val
			} else {
				data.Values[col.Name] = nil
			}
		}
		list = append(list, data)
	}
	return &pb.TableResult{
		Cols:     headers,
		Data:     list,
		Interval: rs.Interval,
	}, nil
}
