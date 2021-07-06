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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
	permission "github.com/erda-project/erda/modules/monitor/common/permission"
)

const permissionResource = "microservice_metric"

func (p *provider) initRoutes(routes httpserver.Router) error {
	const apiPathPrefix = "/api/msp"

	checkByTerminusKeys := permission.Intercepter(
		permission.ScopeProject, p.MPerm.TerminusKeyToProjectIDForHTTP(
			"filter_tk", "filter_terminus_key",
			"filter_target_terminus_key", "filter_source_terminus_key", "filter__metric_scope_id",
		),
		permissionResource, permission.ActionGet,
	)
	routes.GET(apiPathPrefix+"/metrics-query", p.metricQueryByQL, checkByTerminusKeys)
	routes.POST(apiPathPrefix+"/metrics-query", p.metricQueryByQL, checkByTerminusKeys)
	routes.GET(apiPathPrefix+"/metrics/:metric", p.metricQuery, checkByTerminusKeys)
	routes.POST(apiPathPrefix+"/metrics/:metric", p.metricQuery, checkByTerminusKeys)
	routes.GET(apiPathPrefix+"/metrics/:metric/:aggregate", p.metricQuery, checkByTerminusKeys)
	routes.POST(apiPathPrefix+"/metrics/:metric/:aggregate", p.metricQuery, checkByTerminusKeys)

	checkByScopeID := permission.Intercepter(
		permission.ScopeProject, p.MPerm.TerminusKeyToProjectIDForHTTP("scopeId"),
		permissionResource, permission.ActionGet,
	)
	routes.GET(apiPathPrefix+"/metric/groups", p.listGroups, checkByScopeID)
	routes.GET(apiPathPrefix+"/metric/groups/:id", p.getGroup, checkByScopeID)
	// TODO: move to block provider
	routes.GET(apiPathPrefix+"/dashboard/blocks", p.listGroups, checkByScopeID)
	routes.GET(apiPathPrefix+"/dashboard/blocks/:id", p.getGroup, checkByScopeID)
	return nil
}

func (p *provider) metricQueryByQL(rw http.ResponseWriter, r *http.Request) interface{} {
	metric := p.getMetricFromSQL(r)
	if len(metric) <= 0 {
		return api.Errors.InvalidParameter("not found metric name")
	}
	param, err := p.getMetricParams(metric, r)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	return p.proxyMonitor("/api/query", param, rw, r)
}

func (p *provider) metricQuery(rw http.ResponseWriter, r *http.Request, params struct {
	Metric    string `param:"metric"`
	Aggregate string `param:"aggregate"`
}) interface{} {
	param, err := p.getMetricParams(params.Metric, r)
	if err != nil {
		return api.Errors.InvalidParameter(err)
	}
	path := "/api/metrics/" + url.PathEscape(params.Metric)
	if len(params.Aggregate) > 0 {
		path = path + "/" + url.PathEscape(params.Aggregate)
	}
	return p.proxyMonitor(path, param, rw, r)
}

func (p *provider) getMetricParams(metric string, r *http.Request) (url.Values, error) {
	params := url.Values{}
	var key, value string
	for _, k := range []string{
		"filter_tk", "filter_terminus_key", "filter_target_terminus_key",
		"filter_source_terminus_key", "filter__metric_scope_id",
	} {
		val := r.URL.Query().Get(k)
		if len(val) > 0 {
			key, value = k, val
			break
		}
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("not found key")
	}
	params[key] = nil

	tkeys := p.getRuntimeTerminusKeys(value)
	appendTerminusKeys := func(prefix, key string) {
		if len(tkeys) == 1 {
			k := prefix + "eq_" + key
			params.Set(k, tkeys[0])
		} else {
			k := prefix + "in_" + key
			for _, tk := range tkeys {
				params.Add(k, tk)
			}
		}
	}

	if (key == "filter_terminus_key" || key == "filter__metric_scope_id") && r.URL.Query().Get("format") == "chartv2" {
		params["filter__metric_scope"] = nil
		var keys []string
		switch {
		case strings.HasPrefix(metric, "application_") && metric != "application_service_node":
			keys = []string{"target_terminus_key", "source_terminus_key"}
		case strings.HasPrefix(metric, "ta_"):
			keys = []string{"tk"}
		case strings.HasPrefix(metric, "jvm_") || strings.HasPrefix(metric, "nodejs_") ||
			metric == "analyzer_alert" || metric == "error_count" || strings.HasPrefix(metric, "docker_container_summary"):
			keys = []string{"terminus_key"}
		default:
			params.Set("filter__metric_scope", "micro_service")
			keys = []string{"_metric_scope_id"}
		}
		var prefix string
		if len(keys) > 1 {
			prefix = "or_"
		}
		for _, key := range keys {
			appendTerminusKeys(prefix, key)
		}
	} else {
		idx := strings.Index(key, "_")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid key %q", key)
		}
		appendTerminusKeys("", key[idx+1:])
	}
	return params, nil
}

func (p *provider) getMetricFromSQL(r *http.Request) (metric string) {
	params := r.URL.Query()
	q := params.Get("q")
	if len(q) > 0 {
		return strings.TrimSpace(getMetricFromSQL(q))
	} else {
		if params.Get("ql") == "influxql:ast" {
			byts, err := ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewReader(byts))
			if err == nil {
				var body struct {
					From []string `json:"from"`
				}
				json.Unmarshal(byts, &body)
				if len(body.From) != 1 {
					return ""
				}
				return strings.TrimSpace(body.From[0])
			}
		}
	}
	return ""
}

var sqlReg = regexp.MustCompile(`((?i)SELECT)\s+(.*)\s+((?i)FROM)\s+([a-zA-Z0-9_,]+)\s*.*`)

func getMetricFromSQL(sql string) string {
	find := sqlReg.FindAllStringSubmatch(sql, -1)
	if len(find) == 1 {
		find := find[0]
		if len(find) > 0 {
			metrics := strings.Split(find[len(find)-1], ",")
			if len(metrics) != 1 {
				return ""
			}
			return metrics[0]
		}
	}
	return ""
}

func (p *provider) listGroups(rw http.ResponseWriter, r *http.Request, params struct {
	ScopeID string `query:"scopeId" validate:"required"`
}) interface{} {
	param := url.Values{}
	param.Set("scopeId", params.ScopeID)
	return p.proxyMonitor("/api/metric/groups", param, rw, r)
}

func (p *provider) getGroup(rw http.ResponseWriter, r *http.Request, params struct {
	ScopeID string `query:"scopeId" validate:"required"`
	ID      string `param:"id" validate:"required"`
}) interface{} {
	param := url.Values{}
	param.Set("scopeId", params.ScopeID)
	return p.proxyMonitor("/api/metric/groups/"+url.PathEscape(params.ID), param, rw, r)
}

func (p *provider) proxyBlocks(rw http.ResponseWriter, r *http.Request, params struct {
	ScopeID string `query:"scopeId" validate:"required"`
}) interface{} {
	param := url.Values{}
	param.Set("scopeId", params.ScopeID)
	return p.proxyMonitor("/api/dashboard/blocks", param, rw, r)
}

func (p *provider) proxyBlock(rw http.ResponseWriter, r *http.Request, params struct {
	ScopeID string `query:"scopeId" validate:"required"`
	ID      string `param:"id" validate:"required"`
}) interface{} {
	param := url.Values{}
	param.Set("scopeId", params.ScopeID)
	return p.proxyMonitor("/api/dashboard/blocks"+url.PathEscape(params.ID), param, rw, r)
}
