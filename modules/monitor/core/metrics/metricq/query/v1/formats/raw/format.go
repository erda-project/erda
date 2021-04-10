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

package raw

import (
	"fmt"

	query "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1"
	"github.com/olivere/elastic"
)

// Formater .
type Formater struct{}

// Format .
func (f *Formater) Format(ctx *query.Context, param string) (interface{}, error) {
	if len(ctx.Req.GroupBy) == 0 {
		data, err := f.aggData(ctx, ctx.Resp.Aggregations)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return f.groupData(0, ctx, ctx.Resp.Aggregations)
}

func (f *Formater) groupData(depth int, ctx *query.Context, aggs elastic.Aggregations) (interface{}, error) {
	if depth >= len(ctx.Req.GroupBy) {
		return f.aggData(ctx, aggs)
	}
	terms, ok := aggs.Terms(ctx.Req.GroupBy[depth].ID)
	if !ok || terms == nil {
		return nil, fmt.Errorf("invalid terms %s", ctx.Req.GroupBy[depth].ID)
	}
	var list []map[string]interface{}
	for _, term := range terms.Buckets {
		data, err := f.groupData(depth+1, ctx, term.Aggregations)
		if err != nil {
			return nil, err
		}
		group := make(map[string]interface{}, 3)
		group["tag"] = term.Key
		group["total"] = term.DocCount
		group["data"] = data
		list = append(list, group)
	}
	return list, nil
}

func (f *Formater) aggData(ctx *query.Context, aggs elastic.Aggregations) (map[string]interface{}, error) {
	req := ctx.Req
	result := make(map[string]interface{})
	if req.Aggregate != nil {
		agg := req.Aggregate
		switch agg.FuncName {
		case "histogram":
			var list []map[string]interface{}
			histogram, ok := aggs.Histogram(req.Aggregate.ID)
			if !ok || histogram == nil {
				return nil, fmt.Errorf("invalid histogram %s", req.Aggregate.ID)
			}
			interval := ctx.Req.Interval
			alignEnd := ctx.Req.AlignEnd
			points := int(ctx.Req.Points)
			for i, bucket := range histogram.Buckets {
				if i+1 < len(histogram.Buckets) {
					ctx.Attributes["next"] = histogram.Buckets[i+1].Aggregations
				} else {
					delete(ctx.Attributes, "next")
				}
				if i == 0 {
					delete(ctx.Attributes, "previous")
				} else {
					ctx.Attributes["previous"] = histogram.Buckets[i-1].Aggregations
				}
				if i < points {
					item := make(map[string]interface{}, 2)
					var t int64
					if alignEnd {
						t = int64((bucket.Key + interval))
					} else {
						t = int64(bucket.Key)
					}
					item["timestamp"] = t
					data := make(map[string]interface{})
					for _, col := range req.Select {
						value, err := col.Function.Handle(ctx, bucket.Aggregations)
						if err != nil {
							return nil, err
						}
						putColumnData(col, data, value)
					}
					item["data"] = data
					list = append(list, item)
				}
			}
			delete(ctx.Attributes, "next")
			delete(ctx.Attributes, "previous")
			result[agg.FuncName+"."+agg.Property.Name] = map[string]interface{}{
				"agg":  agg.FuncName,
				"name": agg.Property.Name,
				"data": list,
			}
		}
	} else {
		for _, col := range req.Select {
			value, err := col.Function.Handle(ctx, aggs)
			if err != nil {
				return nil, err
			}
			putColumnData(col, result, value)
		}
	}
	return result, nil
}

func putColumnData(col *query.Column, out map[string]interface{}, value interface{}) map[string]interface{} {
	data := make(map[string]interface{}, 3)
	data["data"] = value
	data["name"] = col.Property.Name
	data["agg"] = col.FuncName
	out[col.FuncName+"."+col.Property.Name] = data
	return data
}

func init() {
	query.RegisterResponseFormater("raw", &Formater{})
}
