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
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/chartmeta"
	query "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1"
)

func (f *Formater) isChartBarReq(ctx *query.Context) bool {
	return ctx.Req.ChartType == "chart:bar" && len(ctx.Req.GroupBy) == 1
}

func (f *Formater) formatChartBarChart(ctx *query.Context, chart *chartmeta.ChartMeta) (interface{}, error) {
	aggs := ctx.Resp.Aggregations
	var title string
	if chart != nil {
		title = chart.Title
	}
	var xdata []interface{}
	terms, ok := aggs.Terms(ctx.Req.GroupBy[0].ID)
	if !ok || terms == nil {
		return map[string]interface{}{
			"xdata": xdata,
			"data": []interface{}{
				make(map[string]interface{}, len(ctx.Req.Select)),
			},
			"title": title,
		}, nil
	}
	values := make([][]interface{}, len(ctx.Req.Select))
	for _, term := range terms.Buckets {
		xdata = append(xdata, term.Key)
		for i, col := range ctx.Req.Select {
			value, err := col.Function.Handle(ctx, term.Aggregations)
			if err != nil {
				return nil, err
			}
			value = setDefaultValue(ctx, value)
			values[i] = append(values[i], value)
		}
	}
	data := make(map[string]interface{}, len(ctx.Req.Select))
	for i, col := range ctx.Req.Select {
		putColumnData(col, data, values[i], "", chart)
	}
	return map[string]interface{}{
		"xdata": xdata,
		"data": []interface{}{
			data,
		},
		"title": title,
	}, nil
}
