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

package metricq

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) initRoutesV1(routes httpserver.Router) {
	// metric query apis
	routes.GET("/api/metrics/:scope", p.queryMetricsV1)
	routes.POST("/api/metrics/:scope", p.queryMetricsV1)
	routes.GET("/api/metrics/:scope/:aggregate", p.queryMetricsV1)
	routes.POST("/api/metrics/:scope/:aggregate", p.queryMetricsV1)

	// chart meta
	routes.GET("/api/charts", p.getCharts)
	routes.GET("/api/chart/meta", p.getCharts)
}

func (p *provider) queryMetricsV1(r *http.Request, param *QueryParams) interface{} {
	if len(param.Query) > 0 || strings.HasPrefix(param.QL, "influxql") {
		// 兼容老的 table sql 模式查询
		return p.queryMetrics(r)
	}
	stmt := getQueryStatement(param.Scope, param.Aggregate, r)
	qlang := "json"
	if r.Method == http.MethodGet {
		qlang = "params"
	}
	if len(param.Format) <= 0 {
		param.Format = "chart"
	}
	resp, err := p.q.QueryWithFormatV1(qlang, stmt, param.Format, api.Language(r))
	if err != nil {
		return api.Errors.Internal(err, stmt)
	}

	if len(resp.Details()) > 0 {
		return resp.Details()
	}
	data := resp.Data
	var times, xdata, title interface{}
	if param.Format == "chart" || param.Format == "chartv2" {
		if d, ok := resp.Data.(map[string]interface{}); ok {
			if _, ok := d["metricData"]; param.Format == "chartv2" && ok {
				return api.Success(data)
			}
			data = d["data"]
			if len(resp.Request().GroupBy) <= 0 {
				if data != nil {
					data = []interface{}{
						data,
					}
				} else {
					data = []interface{}{}
				}
			}
			times = d["times"]
			xdata = d["xdata"]
			title = d["title"]
		}
	}

	result := map[string]interface{}{
		"title":    title,
		"total":    resp.Total,
		"interval": resp.Interval,
		"results": []interface{}{
			map[string]interface{}{
				"name": resp.Request().Name,
				"data": data,
			},
		},
	}
	if times != nil {
		result["time"] = times
	} else if xdata != nil {
		result["xData"] = xdata
	}
	return api.Success(result)
}

func getQueryStatement(name, agg string, r *http.Request) string {
	path := name
	if len(agg) > 0 {
		path += "/" + agg
	}
	if r.Method == http.MethodGet {
		return fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
	}
	body, _ := ioutil.ReadAll(r.Body)
	return fmt.Sprintf("%s?%s", path, bytes.NewBuffer(body).String())
}

func (p *provider) getCharts(r *http.Request, param *struct {
	Type string `form:"type" validate:"required"`
}) interface{} {
	list := p.q.Charts(api.Language(r), param.Type)
	var result []interface{}
	// 兼容处理
	for _, item := range list {
		result = append(result, map[string]interface{}{
			"type":       item.Type,
			"name":       item.Name,
			"title":      item.Title,
			"metricName": item.MetricNames,
			"fields":     item.Defines,
			"parameters": item.Parameters,
		})
	}
	return api.Success(result)
}
