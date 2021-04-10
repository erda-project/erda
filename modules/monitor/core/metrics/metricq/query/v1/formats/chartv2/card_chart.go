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

func (f *Formater) isCardReq(ctx *query.Context) bool {
	return len(ctx.Req.GroupBy) == 0
}

func (f *Formater) formatCardChart(ctx *query.Context, chart *chartmeta.ChartMeta) (interface{}, error) {
	aggs := ctx.Resp.Aggregations
	datas := make([]map[string]interface{}, 0)
	for _, col := range ctx.Req.Select {
		value, err := col.Function.Handle(ctx, aggs)
		if err != nil {
			return nil, err
		}
		value = setDefaultValue(ctx, value)
		key := col.FuncName + "." + col.Property.Name
		data := map[string]interface{}{
			"name":  key,
			"value": value,
			// status?: 'rise | fall', // 值的后置图标， 可选上升或下降
			// color?: 'error | cancel | info | success | warning',  // 值的颜色
		}
		if chart != nil && chart.Defines != nil {
			meta := chart.Defines[key]
			if meta != nil {
				if meta.Label != nil && len(*meta.Label) > 0 {
					data["name"] = *meta.Label
				}
				if meta.Unit != nil {
					data["unit"] = *meta.Unit
					var OriginalUnit string
					if meta.OriginalUnit != nil {
						OriginalUnit = *meta.OriginalUnit
					}
					data["value"] = units.Convert(OriginalUnit, *meta.Unit, value)
				}
			}
		}
		datas = append(datas, data)
	}
	return map[string]interface{}{
		"metricData": datas,
	}, nil
}
