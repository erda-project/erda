// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugins

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

type ErdaSBACConfig struct {
	MatchRoute       string   `json:"_match_route_,omitempty" yaml:"_match_route_,omitempty"`
	AccessControlAPI string   `json:"access_control_api,omitempty" yaml:"access_control_api,omitempty"`
	HttpMethods      []string `json:"http_methods,omitempty" yaml:"http_methods,omitempty"`
	MatchPatterns    []string `json:"match_patterns,omitempty" yaml:"match_patterns,omitempty"`
	WithHeaders      []string `json:"with_headers,omitempty" yaml:"with_headers,omitempty"`
	WithCookie       bool     `json:"with_cookie" yaml:"with_cookie"`
	Disable          bool
}

var MSE_ERDA_SBAC_DEFALUT_ACL_LIST = []string{"10.10.10.10", "11.12.13.0/24"}

/* erda-sbac 插件配置格式示例
# 配置必须字段的校验，如下例所示，要求插件配置必须存在 "name"、"age"、“friends” 字段
_rules_:
  - _match_route_:
      - route-erda-default
    access_control_api: "http://test-sbac.default.svc.cluster.local:8080/"
    http_methods:
      - GET
      - PUT
    match_patterns:
      - "^/"
    with_headers:
      - "*"
    with_cookie: true
    with_raw_body: true
*/

func mergeErdaSBACConfig(currentErdaSBACConfig, updateErdaSBACConfig mseDto.MsePluginConfig, updateForDisable bool) (mseDto.MsePluginConfig, error) {
	configBytes, _ := yaml.Marshal(&currentErdaSBACConfig)
	logrus.Debugf("Current ErdaSBAC config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))
	configBytes, _ = yaml.Marshal(&updateErdaSBACConfig)
	logrus.Debugf("Update ErdaSBAC config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))

	// erda-sbac 的 路由唯一性
	// 当前配置项转换
	mapCurrentErdaSBACConfig := getErdaSBACRouteNameToMatchRouteMap(currentErdaSBACConfig, true, updateForDisable)
	logrus.Debugf("mapCurrentErdaSBACConfig=%+v", mapCurrentErdaSBACConfig)

	// 更新配置项转换
	mapUpdateErdaSBACConfig := getErdaSBACRouteNameToMatchRouteMap(updateErdaSBACConfig, false, updateForDisable)
	logrus.Debugf("mapUpdateErdaSBACConfig=%+v", mapUpdateErdaSBACConfig)

	// 更新 routes
	for route, erdaSBACConfig := range mapUpdateErdaSBACConfig {
		if _, ok := mapCurrentErdaSBACConfig[route]; ok {
			// case 1: 如果存在,则判断当前更新是否是删除，是则删除，不是则直接更新
			if erdaSBACConfig.Disable {
				delete(mapCurrentErdaSBACConfig, route)
			} else {
				mapCurrentErdaSBACConfig[route] = erdaSBACConfig
			}
		} else {
			// case 2: 如果不存在，则当前更新时新增，则加入
			if !erdaSBACConfig.Disable {
				mapCurrentErdaSBACConfig[route] = erdaSBACConfig
			}
		}
	}

	rules := make([]mseDto.Rules, 0)
	if len(mapCurrentErdaSBACConfig) == 0 {
		mapCurrentErdaSBACConfig[MseDefaultRouteName] = ErdaSBACConfig{
			MatchRoute:       MseDefaultRouteName,
			AccessControlAPI: common.MseErdaSBACAccessControlAPI,
			HttpMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodConnect,
				http.MethodOptions,
				http.MethodTrace,
			},
			MatchPatterns: []string{common.MseErdaSBACConfigDefaultMatchPattern},
			WithHeaders:   []string{"*"},
			WithCookie:    false,
			Disable:       false,
		}
	}
	for route, erdaSBACConfig := range mapCurrentErdaSBACConfig {
		rule := mseDto.Rules{
			MatchRoute:       []string{route},
			AccessControlAPI: erdaSBACConfig.AccessControlAPI,
			HttpMethods:      erdaSBACConfig.HttpMethods,
			MatchPatterns:    erdaSBACConfig.MatchPatterns,
			WithHeaders:      erdaSBACConfig.WithHeaders,
			WithCookie:       erdaSBACConfig.WithCookie,
		}
		rules = append(rules, rule)
	}

	result := mseDto.MsePluginConfig{
		Rules: rules,
	}

	return result, nil
}

func getErdaSBACRouteNameToMatchRouteMap(pluginConfig mseDto.MsePluginConfig, isCurrentConfig bool, isUpdateForDelete bool) map[string]ErdaSBACConfig {
	routeNameToConfig := make(map[string]ErdaSBACConfig)

	for _, rule := range pluginConfig.Rules {
		for _, route := range rule.MatchRoute {
			erdaSBACConfig := ErdaSBACConfig{
				MatchRoute:       route,
				AccessControlAPI: rule.AccessControlAPI,
				HttpMethods:      rule.HttpMethods,
				MatchPatterns:    rule.MatchPatterns,
				WithHeaders:      rule.WithHeaders,
				WithCookie:       rule.WithCookie,
			}
			if isCurrentConfig {
				erdaSBACConfig.Disable = false
			} else {
				erdaSBACConfig.Disable = isUpdateForDelete
			}
			routeNameToConfig[route] = erdaSBACConfig
		}
	}

	return routeNameToConfig
}

func getErdaSBACSourceConfig(config map[string]interface{}) (accessControlAPI string, matchPatterns []string, httpMethods []string, withHeaders []string, withCookie bool, disable bool, err error) {

	withCookie = false

	api, ok := config[common.MseErdaSBACConfigAccessControlAPI]
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("not set access_control_api for plugin %s", common.MsePluginSbac)
	}
	accessControlAPI, ok = api.(string)
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("access_control_api for plugin %s is invalid string", common.MsePluginSbac)
	}

	patterns, ok := config[common.MseErdaSBACConfigMatchPatterns]
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("not set patterns for plugin %s", common.MsePluginSbac)
	}
	matchPatterns, ok = patterns.([]string)
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("patterns for plugin %s is invalid []string", common.MsePluginSbac)
	}

	methods, ok := config[common.MseErdaSBACConfigHttpMethods]
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("not set methods for plugin %s", common.MsePluginSbac)
	}
	methodsMap, ok := methods.(map[string]bool)
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("methods for plugin %s is invalid map[string]bool", common.MsePluginSbac)
	}
	httpMethods = make([]string, 0)
	for method := range methodsMap {
		httpMethods = append(httpMethods, method)
	}

	headers, ok := config[common.MseErdaSBACConfigWithHeaders]
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("not set with_headers for plugin %s", common.MsePluginSbac)
	}

	withHeaders, ok = headers.([]string)
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("with_headers for plugin %s is invalid []string", common.MsePluginSbac)
	}

	withCookie = false
	cookie, ok := config[common.MseErdaSBACConfigWithCookie]
	if ok {
		withCookie, ok = cookie.(bool)
		if !ok {
			return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("with_cookie for plugin %s is invalid []string", common.MsePluginSbac)
		}
	}

	routeSwitch, ok := config[common.MseErdaSBACRouteSwitch]
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("not set %s for plugin %s", common.MseErdaSBACRouteSwitch, common.MsePluginSbac)
	}
	enable, ok := routeSwitch.(bool)
	if !ok {
		return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, false, errors.Errorf("%s for plugin %s is invalid bool", common.MseErdaSBACRouteSwitch, common.MsePluginSbac)
	}

	disable = !enable

	return accessControlAPI, matchPatterns, httpMethods, withHeaders, withCookie, disable, nil
}
