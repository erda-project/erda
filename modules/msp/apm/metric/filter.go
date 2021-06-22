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
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type TSQLQueryRequest struct {
	From []string `json:"from"`
}

type cluster struct {
	clusterName string `gorm:"column:cluster_name"`
}

func (m *provider) proxy(metric, agg string, request *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(request.Body)
	bodyReader := bytes.NewReader(body)
	if err != nil {
		return nil, err
	}
	if metric == "" && agg == "" {
		q := request.URL.Query().Get("q")
		if q != "" {
			// "SELECT * FROM index WHERE ..."
			pattern := "((?i)SELECT)\\s+(.*)\\s+((?i)FROM)\\s+([a-zA-Z0-9_]+)\\s*.*"
			r, err := regexp.Compile(pattern)
			if err != nil {
				return nil, err
			}
			m := r.FindStringSubmatch(metric)
			if len(m) >= 5 {
				metric = m[4]
			}
		} else {
			var m TSQLQueryRequest
			err = json.Unmarshal(body, &m)
			if err != nil {
				return nil, err
			}
			if m.From != nil && len(m.From) > 0 {
				metric = m.From[0]
			}
		}
		if metric == "" {
			return nil, errors.New("metric not present")
		}
	}
	var params = make(map[string][]string)
	var terminusKey string
	values := getRequestValue(request,
		"filter_tk", "filter_terminus_key", "filter_target_terminus_key", "filter_source_terminus_key", "filter__metric_scope_id")
	if len(values) > 0 {
		terminusKey = values[1]
		key := values[0]
		convertTerminusKeys(metric, key, terminusKey, params, request.URL.Query().Get("format") == "chartv2")
	}

	if !strings.HasPrefix(metric, "ta_") {
		values = getRequestValue(request, "filter_cluster_name")
		if len(values) == 0 {
			if terminusKey != "" {
				var clusters []cluster
				err := m.db.Raw("SELECT DISTINCT cluster_name FROM sp_monitor WHERE terminus_key=? AND is_delete = 0", terminusKey).Scan(&clusters).Error
				if err != nil {
					return nil, err
				}
				if len(clusters) > 0 {
					for _, cluster := range clusters {
						params["in_cluster_name"] = append(params["in_cluster_name"], cluster.clusterName)
					}
				}
			} else {
				values = getRequestValue(request, "filter_project_id")
				if len(values) > 0 {
					var clusters []string
					err := m.db.Raw("SELECT DISTINCT cluster_name FROM sp_monitor WHERE project_id=? AND is_delete = 0", values[1]).Find(&clusters).Error
					if err != nil {
						return nil, err
					}
					if len(clusters) > 0 {
						{
							params["in_cluster_name"] = append(params["in_cluster_name"], clusters...)
						}
					}
				}
			}
		}
	}
	var path string
	if agg == "" {
		path = m.Cfg.MonitorServiceMetricApiPath + "/" + metric
	} else {
		path = m.Cfg.MonitorServiceMetricApiPath + "/" + metric + "/" + agg
	}

	return m.Proxy(path, request, params, bodyReader, request.URL.Query().Get("debug") != "")
}

func (m *provider) proxyBlocks(scopeId, id string, request *http.Request) (interface{}, error) {
	var params = make(map[string][]string)
	params["scope"] = append(params["scope"], "micro_service")
	params["scopeId"] = append(params["scopeId"], scopeId)
	if id == "" {
		return m.ProxyBody("/api/dashboard/blocks", request, params, request.URL.Query().Get("debug") != "")
	}
	return m.ProxyBody("/api/dashboard/blocks/"+id, request, params, request.URL.Query().Get("debug") != "")
}

func (m *provider) proxyGroups(scopeId, id string, request *http.Request) (interface{}, error) {
	var params = make(map[string][]string)
	params["scope"] = append(params["scope"], "micro_service")
	params["scopeId"] = append(params["scopeId"], scopeId)
	if id == "" {
		return m.ProxyBody("/api/metric/groups", request, params, request.URL.Query().Get("debug") != "")
	}
	return m.ProxyBody("/api/metric/groups/"+id, request, params, request.URL.Query().Get("debug") != "")
}

func getRequestValue(request *http.Request, keys ...string) []string {
	vars := request.URL.Query()
	for _, key := range keys {
		value := vars[key]
		if len(value) > 0 {
			return []string{key, value[0]}
		}
	}
	return nil
}

func convertTerminusKeys(metric, keyName, tk string, params map[string][]string, rawQuery bool) map[string][]string {
	params[keyName] = nil
	tkList := getRuntimeTerminusKeys(tk)
	if (keyName == "filter_terminus_key" || keyName == "filter__metric_scope_id") && rawQuery {
		params["filter__metric_scope"] = nil
		var keys = make(map[string]string)
		if strings.HasPrefix(metric, "application_") && metric != "application_service_node" {
			keys["target_terminus_key"] = ""
			keys["source_terminus_key"] = ""
		} else if strings.HasPrefix(metric, "ta_") {
			keys["tk"] = ""
		} else if strings.HasPrefix(metric, "jvm_") || strings.HasPrefix(metric, "nodejs_") {
			keys["terminus_key"] = ""
		} else if metric == "analyzer_alert" || metric == "error_count" {
			keys["terminus_key"] = ""
		} else if strings.HasPrefix(metric, "docker_container_summary") {
			// do nothing
		} else {
			params["filter__metric_scope"] = append(params["filter__metric_scope"], "micro_service")
			keys["_metric_scope_id"] = ""
		}
		prefix := ""
		if len(keys) > 0 {
			prefix = "or_"
		}
		for key := range keys {
			appendTerminusKeys(prefix, key, tkList, params)
		}
	} else {

		idx := strings.IndexAny(keyName, "_")
		appendTerminusKeys("", keyName[:idx+1], tkList, params)
	}
	return params
}

func appendTerminusKeys(prefix, key string, values []string, params map[string][]string) {
	if len(values) > 0 {
		if len(values) == 1 {
			key = prefix + "eq_" + key
			delete(params, key)
			params[key] = append(params[key], values[0])
		} else {
			k := prefix + "in_" + key
			delete(params, k)
			for _, tk := range values {
				if tk != "" {
					params[k] = append(params[k], tk)
				}
			}
		}
	}
}

var tkMap = make(map[string][]string)

func getRuntimeTerminusKeys(terminusKey string) []string {
	var result []string
	keys := tkMap[terminusKey]
	if len(keys) == 0 {
		for _, key := range keys {
			if key != "" && !(key == terminusKey) {
				result = append(result, key)
			}
		}
	}
	result = append(result, terminusKey)
	return result
}
