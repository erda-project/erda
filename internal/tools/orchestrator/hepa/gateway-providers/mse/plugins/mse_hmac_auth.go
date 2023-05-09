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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

func mergeHmacAuthConfig(CurrentHmacAuthConfig, updateHmacAuthConfig mseDto.MsePluginConfig) (mseDto.MsePluginConfig, error) {
	configBytes, _ := yaml.Marshal(&CurrentHmacAuthConfig)
	logrus.Debugf("Current HmacAuth config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))
	configBytes, _ = yaml.Marshal(&updateHmacAuthConfig)
	logrus.Debugf("Update HmacAuth config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))

	// 当前配置项转换
	// Hmac-auth 的 key 是唯一性的
	mapCurrentHmacAuthKeyToConsumer, mapCurrentHmacAuthRouteToConsumers := getConsumersAndRoutesMaps(CurrentHmacAuthConfig.Consumers, CurrentHmacAuthConfig.Rules)

	// 更新配置项转换
	mapUpdateHmacAuthKeyToConsumer, mapUpdateHmacAuthRouteToConsumers := getConsumersAndRoutesMaps(updateHmacAuthConfig.Consumers, updateHmacAuthConfig.Rules)

	// 1. 更新 consumers (加入)
	for key, hmacAuthConsumer := range mapUpdateHmacAuthKeyToConsumer {
		currentHmacAuthConsumer, ok := mapCurrentHmacAuthKeyToConsumer[key]
		if !ok {
			mapCurrentHmacAuthKeyToConsumer[key] = hmacAuthConsumer
		} else {
			if currentHmacAuthConsumer.Name != hmacAuthConsumer.Name {
				return CurrentHmacAuthConfig, errors.Errorf("existed consumer %s has the same key with new consumer %s\n", currentHmacAuthConsumer.Name, hmacAuthConsumer.Name)
			}
		}
	}

	// 2. 更新 routes
	for route, updateConsumers := range mapUpdateHmacAuthRouteToConsumers {
		if currentConsumers, existed := mapCurrentHmacAuthRouteToConsumers[route]; existed {
			// routes 里有consumer 的信息，则要检查 consumer 对应的 routes 的列表中是否已经包含了当前 routes，没包含则加入，有包含则去重之后再加入
			for _, cs := range updateConsumers {
				if !isInList(currentConsumers, cs) {
					mapCurrentHmacAuthRouteToConsumers[route] = append(mapCurrentHmacAuthRouteToConsumers[route], cs)
				}
			}
		} else {
			// 没有 route，则直接加入
			mapCurrentHmacAuthRouteToConsumers[route] = updateConsumers
		}
	}

	// 3. 删除 route 不应该有权限的 consumer 及 allows
	routesConsumersToDel := make(map[string][]string)
	for route, updateConsumers := range mapUpdateHmacAuthRouteToConsumers {
		currentConsumers, ok := mapCurrentHmacAuthRouteToConsumers[route]
		if ok {
			for _, cs := range currentConsumers {
				if !isInList(updateConsumers, cs) {
					if _, exist := routesConsumersToDel[route]; !exist {
						routesConsumersToDel[route] = make([]string, 0)
					}
					routesConsumersToDel[route] = append(routesConsumersToDel[route], cs)
				}
			}
		}
	}
	for route, consumerNames := range routesConsumersToDel {
		mapCurrentHmacAuthRouteToConsumers[route] = deleteSubListFromList(mapCurrentHmacAuthRouteToConsumers[route], consumerNames)
		if len(mapCurrentHmacAuthRouteToConsumers[route]) == 0 {
			delete(mapCurrentHmacAuthRouteToConsumers, route)
		}
	}

	// 4. 更新 consumers (如果某个 consumer 对应已经没有路由配置，则移除 consumer)
	keysToDelete := make([]string, 0)
	for key, cs := range mapCurrentHmacAuthKeyToConsumer {
		used := false
		for _, css := range mapCurrentHmacAuthRouteToConsumers {
			if isInList(css, cs.Name) {
				used = true
				break
			}
		}
		if used {
			continue
		}
		keysToDelete = append(keysToDelete, key)
	}

	for _, key := range keysToDelete {
		delete(mapCurrentHmacAuthKeyToConsumer, key)
	}

	// 如果 mapCurrentHmacAuthKeyToConsumer  中不包含默认的 consumer（默认的路由用于确保插件配置不会成为 全局配置），则添加
	if _, ok := mapCurrentHmacAuthKeyToConsumer[MseDefaultConsumerCredential]; !ok {
		mapCurrentHmacAuthKeyToConsumer[MseDefaultConsumerCredential] = KeySecretConsumer{
			Name:   MseDefaultConsumerName,
			Key:    MseDefaultConsumerKey,
			Secret: MseDefaultConsumerSecret,
		}
	}

	consumers := make([]mseDto.Consumers, 0)
	for key, cs := range mapCurrentHmacAuthKeyToConsumer {
		consumers = append(consumers, mseDto.Consumers{
			Name:   cs.Name,
			Key:    key,
			Secret: cs.Secret,
		})
	}

	// 如果 mapCurrentHmacAuthRouteToConsumers 中不包含默认的路由（默认的路由用于确保插件配置不会成为 全局配置），则添加
	if _, ok := mapCurrentHmacAuthRouteToConsumers[MseDefaultRouteName]; !ok {
		mapCurrentHmacAuthRouteToConsumers[MseDefaultRouteName] = []string{MseDefaultConsumerName}
	}

	rules := make([]mseDto.Rules, 0)
	for route, cs := range mapCurrentHmacAuthRouteToConsumers {
		// 删除非默认的路由对应的只包含默认 Consumer 的  routes 项（这种一般表示路由无授权用户或者路由被从前端页面删除了）
		if route != MseDefaultRouteName && len(cs) == 1 && cs[0] == MseDefaultConsumerName {
			continue
		}
		rules = append(rules, mseDto.Rules{
			MatchRoute: []string{route},
			Allow:      cs,
		})
	}

	result := mseDto.MsePluginConfig{
		Consumers: consumers,
		Rules:     rules,
	}

	return result, nil
}

func getConsumersAndRoutesMaps(consumers []mseDto.Consumers, rules []mseDto.Rules) (map[string]KeySecretConsumer, map[string][]string) {
	mapKeyToConsumer := make(map[string]KeySecretConsumer)
	mapRouteToConsumers := make(map[string][]string)

	for _, cs := range consumers {
		mapKeyToConsumer[cs.Key] = KeySecretConsumer{
			Name:   cs.Name,
			Key:    cs.Key,
			Secret: cs.Secret,
		}
	}

	// []Rules 部分 使用 map[route][]consumers 的方式构建，便于比较，另外更新后的策略也是这种 单个用户对多条路由的情况
	for _, rule := range rules {
		for _, route := range rule.MatchRoute {
			if _, ok := mapRouteToConsumers[route]; !ok {
				mapRouteToConsumers[route] = make([]string, 0)
			}
			mapRouteToConsumers[route] = append(mapRouteToConsumers[route], rule.Allow...)
		}
	}

	return mapKeyToConsumer, mapRouteToConsumers
}
