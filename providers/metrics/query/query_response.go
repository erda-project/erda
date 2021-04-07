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

package query

import (
	"encoding/json"

	"github.com/jmespath/go-jmespath"
	"github.com/mitchellh/mapstructure"
)

type MetricQueryResponse struct {
	StatusCode int
	Body       []byte
}

// 返回多值时序数据
func (resp *MetricQueryResponse) ReturnAsSeries() (out *Series, err error) {
	var all interface{}
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		return nil, err
	}
	out = &Series{
		TimeSeries: make([]int, 0),
		Data:       make([]*SeriesData, 0),
	}

	if v, err := jmespath.Search("data.results[0].data[0]", all); err != nil {
		return nil, err
	} else {
		tmp := v.(map[string]interface{})
		for _, val := range tmp {
			var p SeriesData
			if err := mapstructure.Decode(val, &p); err != nil {
				return nil, err
			}
			out.Data = append(out.Data, &p)
		}
	}

	if v, err := jmespath.Search("data.time", all); err != nil {
		return nil, err
	} else {
		tmp := v.([]interface{})
		for _, val := range tmp {
			out.TimeSeries = append(out.TimeSeries, int(val.(float64)))
		}
	}

	if v, err := jmespath.Search("data.results[0].name", all); err != nil {
		return nil, err
	} else {
		out.Name = v.(string)
	}

	return out, nil
}

// 返回单值数据
func (resp *MetricQueryResponse) ReturnAsPoint() (out *Point, err error) {
	var all interface{}
	if err := json.Unmarshal(resp.Body, &all); err != nil {
		return nil, err
	}
	out = &Point{
		Data: make([]*PointData, 0),
	}

	if v, err := jmespath.Search("data.results[0].data[0]", all); err != nil {
		return nil, err
	} else {
		tmp := v.(map[string]interface{})
		for _, val := range tmp {
			var p PointData
			if err := mapstructure.Decode(val, &p); err != nil {
				return nil, err
			}
			out.Data = append(out.Data, &p)
		}
	}

	if v, err := jmespath.Search("data.results[0].name", all); err != nil {
		return nil, err
	} else {
		out.Name = v.(string)
	}

	return out, nil
}
