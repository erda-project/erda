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
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

type ParaSignConfig struct {
	MatchRoute           string             `json:"_match_route_,omitempty" yaml:"_match_route_,omitempty"`
	Consumers            []mseDto.Consumers `json:"consumers,omitempty" yaml:"consumers,omitempty"`
	RequestBodySizeLimit int                `json:"request_body_size_limit,omitempty" yaml:"request_body_size_limit,omitempty"`
	DateOffset           int                `json:"date_offset,omitempty" yaml:"date_offset,omitempty"`
}

/* para-sign-auth 插件配置格式示例
_rules_:
- _match_route_:
  - route-erda-default
  request_body_size_limit: 10485760
  date_offset: 600
  consumers:
  - name: consumer-erda-default
    key: 2bda943c-ba2b-11ec-ba07-00163e1250b5
    secret: 2bda943c-ba2b-11ec-ba07-00163e1250b5
*/

func mergeParaSignAuthConfig(currentParaSignAuthConfig, updateParaSignAuthConfig mseDto.MsePluginConfig) (mseDto.MsePluginConfig, error) {
	configBytes, _ := yaml.Marshal(&currentParaSignAuthConfig)
	logrus.Debugf("Current ParaSignAuth config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))
	configBytes, _ = yaml.Marshal(&updateParaSignAuthConfig)
	logrus.Debugf("Update ParaSignAuth config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))

	// Para-sign-auth 的 路由唯一性的
	// 当前配置项转换
	mapCurrentParaSignConfig := getParaSignRouteNameToMatchRouteMap(currentParaSignAuthConfig)
	logrus.Debugf("mapCurrentParaSignConfig=%+v", mapCurrentParaSignConfig)

	// 更新配置项转换
	mapUpdateParaSignConfig := getParaSignRouteNameToMatchRouteMap(updateParaSignAuthConfig)
	logrus.Debugf("mapUpdateParaSignConfig=%+v", mapUpdateParaSignConfig)

	// 更新 routes
	for route, paraSignConfig := range mapUpdateParaSignConfig {
		currentParaSignConfig, ok := mapCurrentParaSignConfig[route]
		if !ok {
			// 1. 现有配置不存在，直接加入
			if paraSignConfig.RequestBodySizeLimit == 0 {
				paraSignConfig.RequestBodySizeLimit = REQUEST_BODY_SIZE_LIMIT
			}
			mapCurrentParaSignConfig[route] = paraSignConfig
		} else {
			// 2. 如果更新为空 consumers ，则直接直接删除，否则进行替换
			if len(paraSignConfig.Consumers) == 0 {
				delete(mapCurrentParaSignConfig, route)
			} else {
				// 不同更新可能会修改 RequestBodySizeLimit， 保留最大值对应的修改
				if paraSignConfig.RequestBodySizeLimit <= currentParaSignConfig.RequestBodySizeLimit {
					paraSignConfig.RequestBodySizeLimit = currentParaSignConfig.RequestBodySizeLimit
				}
				// 不同更新可能会修改 DateOffset， 保留最大值对应的修改
				if paraSignConfig.DateOffset > 0 && paraSignConfig.DateOffset < currentParaSignConfig.DateOffset {
					paraSignConfig.DateOffset = currentParaSignConfig.DateOffset
				}
				mapCurrentParaSignConfig[route] = paraSignConfig
			}
		}
	}

	rules := make([]mseDto.Rules, 0)
	for route, paraSignConfig := range mapCurrentParaSignConfig {
		// 删除非默认的路由对应的只包含默认 Consumer 的  routes 项（这种一般表示路由无授权用户或者路由被从前端页面删除了）
		if route != DEFAULT_MSE_ROUTE_NAME && len(paraSignConfig.Consumers) == 1 &&
			paraSignConfig.Consumers[0].Key == DEFAULT_MSE_CONSUMER_KEY &&
			paraSignConfig.Consumers[0].Name == DEFAULT_MSE_CONSUMER_NAME &&
			paraSignConfig.Consumers[0].Secret == DEFAULT_MSE_CONSUMER_SECRET {
			continue
		}

		rule := mseDto.Rules{
			MatchRoute: []string{route},
			Consumers:  paraSignConfig.Consumers,
		}

		if paraSignConfig.RequestBodySizeLimit > 0 {
			rule.RequestBodySizeLimit = paraSignConfig.RequestBodySizeLimit
		} else {
			rule.RequestBodySizeLimit = REQUEST_BODY_SIZE_LIMIT
		}

		if paraSignConfig.DateOffset > 0 {
			rule.DateOffset = paraSignConfig.DateOffset
		}

		rules = append(rules, rule)
	}

	result := mseDto.MsePluginConfig{
		Rules: rules,
	}

	return result, nil
}

func getParaSignRouteNameToMatchRouteMap(pluginConfig mseDto.MsePluginConfig) map[string]ParaSignConfig /*(map[string]ParaSignConfig , map[string][]string) */ {
	routeNameToConfig := make(map[string]ParaSignConfig)

	for _, rule := range pluginConfig.Rules {
		for _, route := range rule.MatchRoute {
			paraSignConfig := ParaSignConfig{
				MatchRoute: route,
				Consumers:  rule.Consumers,
			}
			if rule.DateOffset > 0 {
				paraSignConfig.DateOffset = rule.DateOffset
			}

			if rule.RequestBodySizeLimit > 0 {
				paraSignConfig.RequestBodySizeLimit = rule.RequestBodySizeLimit
			} else {
				paraSignConfig.RequestBodySizeLimit = REQUEST_BODY_SIZE_LIMIT
			}

			if rule.DateOffset > 0 {
				paraSignConfig.DateOffset = rule.DateOffset
			}
			routeNameToConfig[route] = paraSignConfig
		}
	}

	return routeNameToConfig
}
