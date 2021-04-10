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
	"io/ioutil"
	"net/http"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) initRoutes(routes httpserver.Router) error {
	p.initRoutesV1(routes)

	// metric query apis
	routes.GET("/query", p.queryMetrics)      // like influxdb query api
	routes.POST("/query", p.queryMetrics)     // like influxdb query api
	routes.GET("/api/query", p.queryMetrics)  // for tsql
	routes.POST("/api/query", p.queryMetrics) // for tsql

	// 数据导出，临时方案
	routes.GET("/api/metrics/:scope/export", p.exportMetrics)
	routes.POST("/api/metrics/:scope/export", p.exportMetrics)

	// metric meta
	routes.GET("/api/metric/names", p.listMetricNames)
	routes.GET("/api/metric/meta", p.listMetricMeta)
	// routes.POST("/api/metric/meta", nil) // 注册指标元数据，TODO .
	routes.GET("/api/metric/groups", p.listMetricGroups)
	routes.GET("/api/metric/groups/:id", p.getMetricGroup)
	// routes.POST("/api/metric/groups", nil) // 注册指标分组信息，TODO .
	routes.GET("/api/metadata/groups", p.listMetricGroups)   // 复用之前的接口路径
	routes.GET("/api/metadata/groups/:id", p.getMetricGroup) // 复用之前的接口路径
	return nil
}

// queryMetrics .
func (p *provider) queryMetrics(r *http.Request) interface{} {
	err := r.ParseForm()
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	ql, q, format := r.Form.Get("ql"), r.Form.Get("q"), r.Form.Get("format")
	r.Form.Del("ql")
	r.Form.Del("q")
	r.Form.Del("format")
	if len(format) == 0 {
		format = "influxdb"
	}
	if len(ql) == 0 {
		ql = "influxql"
	}
	if len(q) == 0 {
		byts, err := ioutil.ReadAll(r.Body)
		if err == nil {
			q = string(byts)
		}
	}
	resp, data, err := p.q.QueryWithFormat(ql, q, format, api.Language(r), nil, r.Form)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	if resp.Details != nil {
		return resp.Details
	}
	if response, ok := data.(httpserver.Response); ok {
		return response
	} else if response, ok := data.(httpserver.ResponseGetter); ok {
		return response
	}
	return api.Success(data)
}
