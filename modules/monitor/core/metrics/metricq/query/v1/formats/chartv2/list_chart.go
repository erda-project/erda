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
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/units"
)

func (f *Formater) isListReq(ctx *query.Context) bool {
	return len(ctx.Req.GroupBy) == 1 && len(ctx.Req.Select) == 1
}

func (f *Formater) formatListChart(ctx *query.Context, chart *chartmeta.ChartMeta) (interface{}, error) {
	aggs := ctx.Resp.Aggregations
	list := make([]map[string]interface{}, 0)
	terms, ok := aggs.Terms(ctx.Req.GroupBy[0].ID)
	if !ok || terms == nil {
		return map[string]interface{}{
			"metricData": list,
		}, nil
	}
	col := ctx.Req.Select[0]
	key := col.FuncName + "." + col.Property.Name

	var meta *chartmeta.DataMeta
	if chart != nil && chart.Defines != nil {
		meta = chart.Defines[key]
	}
	for _, term := range terms.Buckets {
		value, err := col.Function.Handle(ctx, term.Aggregations)
		if err != nil {
			return nil, err
		}
		value = setDefaultValue(ctx, value)
		data := map[string]interface{}{
			"title": term.Key,
			"value": value,
		}
		if ctx.Req.TransGroup {
			if key, ok := term.Key.(string); ok {
				data["title"] = ctx.T.Text(ctx.Lang, key)
			}
		} else if title, ok := term.Key.(string); ok {
			if chart != nil && chart.Defines != nil && chart.Defines[title] != nil && chart.Defines[title].Label != nil {
				data["title"] = chart.Defines[title].Label
			}
		}
		if meta != nil {
			if meta.Unit != nil {
				data["unit"] = meta.Unit
			}
			if meta.Unit != nil {
				var OriginalUnit string
				if meta.OriginalUnit != nil {
					OriginalUnit = *meta.OriginalUnit
				}
				data["value"] = units.Convert(OriginalUnit, *meta.Unit, value)
			}
		}
		list = append(list, data)
	}
	return map[string]interface{}{
		"metricData": list,
	}, nil
}
