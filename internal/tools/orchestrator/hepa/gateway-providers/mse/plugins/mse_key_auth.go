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

func mergeKeyAuthConfig(CurrentKeyAuthConfig, updateKeyAuthConfig mseDto.MsePluginConfig) (mseDto.MsePluginConfig, error) {
	configBytes, _ := yaml.Marshal(&CurrentKeyAuthConfig)

	logrus.Debugf("Current KeyAuth config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))
	configBytes, _ = yaml.Marshal(&updateKeyAuthConfig)
	logrus.Debugf("Update KeyAuth config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))

	// 当前配置项转换
	// Key-auth 的 Credential 是唯一性的（但 name 不唯一，因为一个 name 可以有多个 credential）
	mapCurrentKeyAuthCredentialToConsumer, mapCurrentKeyAuthRouteToConsumers := getCredentialsAndRoutesMaps(CurrentKeyAuthConfig.Consumers, CurrentKeyAuthConfig.Rules)

	// 更新配置项转换
	mapUpdateKeyAuthCredentialToConsumer, mapUpdateKeyAuthRouteToConsumers := getCredentialsAndRoutesMaps(updateKeyAuthConfig.Consumers, updateKeyAuthConfig.Rules)

	// 1. 更新 consumers (加入)
	for credential, name := range mapUpdateKeyAuthCredentialToConsumer {
		csName, ok := mapCurrentKeyAuthCredentialToConsumer[credential]
		if !ok {
			mapCurrentKeyAuthCredentialToConsumer[credential] = name
		} else {
			if csName != name {
				return CurrentKeyAuthConfig, errors.Errorf("existed consumer %s has the same credential with new consumer %s\n", csName, name)
			}
		}
	}

	// 2. 更新 routes
	for route, updateConsumers := range mapUpdateKeyAuthRouteToConsumers {
		if currentConsumers, existed := mapCurrentKeyAuthRouteToConsumers[route]; existed {
			// routes 里有consumer 的信息，则要检查 consumer 对应的 routes 的列表中是否已经包含了当前 routes，没包含则加入，有包含则去重之后再加入
			for _, cs := range updateConsumers {
				if !isInList(currentConsumers, cs) {
					mapCurrentKeyAuthRouteToConsumers[route] = append(mapCurrentKeyAuthRouteToConsumers[route], cs)
				}
			}
		} else {
			// 没有 route，则直接加入
			mapCurrentKeyAuthRouteToConsumers[route] = updateConsumers
		}
	}

	// 3. 删除 route 不应该有权限的 consumer 及 allows
	routesConsumersToDel := make(map[string][]string)
	for route, updateConsumers := range mapUpdateKeyAuthRouteToConsumers {
		currentConsumers, ok := mapCurrentKeyAuthRouteToConsumers[route]
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
		mapCurrentKeyAuthRouteToConsumers[route] = deleteSubListFromList(mapCurrentKeyAuthRouteToConsumers[route], consumerNames)
		if len(mapCurrentKeyAuthRouteToConsumers[route]) == 0 {
			delete(mapCurrentKeyAuthRouteToConsumers, route)
		}
	}

	// 4. 更新 consumers (如果某个 consumer 对应已经没有路由配置，则移除 consumer)
	credentialsToDelete := make([]string, 0)
	for credential, csName := range mapCurrentKeyAuthCredentialToConsumer {
		used := false
		for _, css := range mapCurrentKeyAuthRouteToConsumers {
			if isInList(css, csName) {
				used = true
				break
			}
		}
		if used {
			continue
		}
		credentialsToDelete = append(credentialsToDelete, credential)
	}

	for _, cred := range credentialsToDelete {
		delete(mapCurrentKeyAuthCredentialToConsumer, cred)
	}

	// 如果 mapCurrentKeyAuthCredentialToConsumer  中不包含默认的consumer（默认的路由用于确保插件配置不会成为 全局配置），则添加
	if _, ok := mapCurrentKeyAuthCredentialToConsumer[DEFAULT_MSE_CONSUMER_CREDENTIAL]; !ok {
		mapCurrentKeyAuthCredentialToConsumer[DEFAULT_MSE_CONSUMER_CREDENTIAL] = DEFAULT_MSE_CONSUMER_NAME
	}
	consumers := make([]mseDto.Consumers, 0)
	for credential, csName := range mapCurrentKeyAuthCredentialToConsumer {
		consumers = append(consumers, mseDto.Consumers{
			Name:       csName,
			Credential: credential,
		})
	}

	// 5. 更新 routes
	// 如果 mapCurrentKeyAuthRouteToConsumers 中不包含默认的路由（默认的路由用于确保插件配置不会成为 全局配置），则添加
	if _, ok := mapCurrentKeyAuthRouteToConsumers[DEFAULT_MSE_ROUTE_NAME]; !ok {
		mapCurrentKeyAuthRouteToConsumers[DEFAULT_MSE_ROUTE_NAME] = []string{DEFAULT_MSE_CONSUMER_NAME}
	}

	rules := make([]mseDto.Rules, 0)
	for route, cs := range mapCurrentKeyAuthRouteToConsumers {
		// 删除非默认的路由对应的只包含默认 Consumer 的  routes 项（这种一般表示路由无授权用户或者路由被从前端页面删除了）
		if route != DEFAULT_MSE_ROUTE_NAME && len(cs) == 1 && cs[0] == DEFAULT_MSE_CONSUMER_NAME {
			continue
		}
		rules = append(rules, mseDto.Rules{
			MatchRoute: []string{route},
			Allow:      cs,
		})
	}

	result := mseDto.MsePluginConfig{
		Consumers: consumers,
		Keys:      []string{"appKey", "x-app-key"},
		InQuery:   true,
		InHeader:  true,
		Rules:     rules,
	}

	return result, nil
}
