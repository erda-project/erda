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
	mseclient "github.com/alibabacloud-go/mse-20190531/v3/client"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

const (
	DEFAULT_MSE_ROUTE_NAME          = "route-erda-default"
	DEFAULT_MSE_CONSUMER_NAME       = "consumer-erda-default"
	DEFAULT_MSE_CONSUMER_CREDENTIAL = "2bda943c-ba2b-11ec-ba07-00163e1250b5"
)

func UpdatePluginConfigWhenDeleteConsumer(pluginName, consumerName string, config interface{}) ([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, error) {
	// 只看全局配置对应的列表的第一个，因为当前(2023.02.23)只支持全局配置,且只会有一条配置记录
	pluginConfig, ok := config.([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList)
	if !ok {
		return nil, errors.Errorf("config is not type []*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList")
	}

	currentConf := ""
	index := -1
	for idx, conf := range pluginConfig {
		// 只看全局配置对应的列表，因为当前(2023.02.23)只支持全局配置
		if *conf.ConfigLevel != 0 {
			continue
		}
		currentConf = *conf.Config
		index = idx
		break
	}

	if currentConf != "" {
		switch pluginName {
		case common.MsePluginKeyAuth:
			keyAutoConfig := mseDto.MsePluginKeyAuthConfig{}
			err := yaml.Unmarshal([]byte(currentConf), &keyAutoConfig)
			if err != nil {
				return nil, err
			}
			mapCredentialToConsumerName, mapConsumerNameToRoutes := updateWithDeleteConsumer(consumerName, keyAutoConfig.Consumers, keyAutoConfig.Rules)

			keyAutoConfig.Consumers = make([]mseDto.Consumers, 0)
			for cred, consumer := range mapCredentialToConsumerName {
				keyAutoConfig.Consumers = append(keyAutoConfig.Consumers, mseDto.Consumers{
					Name:       consumer,
					Credential: cred,
				})
			}

			keyAutoConfig.Rules = make([]mseDto.Rules, 0)
			for allow, routes := range mapConsumerNameToRoutes {
				keyAutoConfig.Rules = append(keyAutoConfig.Rules, mseDto.Rules{
					MatchRoute: routes,
					Allow:      []string{allow},
				})
			}

			configBytes, err := yaml.Marshal(&keyAutoConfig)
			if err != nil {
				return nil, err
			}

			currentConf = string(configBytes)
			logrus.Infof("Yaml file content: \n%s\n", string(configBytes))
			pluginConfig[index].Config = &currentConf

		case common.MsePluginBasicAuth:
		case common.MsePluginHmacAuth:
		case common.MsePluginCustomResponse:
		case common.MsePluginRequestBlock:
		case common.MsePluginBotDetect:
		case common.MsePluginKeyRateLimit:
		case common.MsePluginHttp2Misdirect:
		case common.MsePluginJwtAuth:
		case common.MsePluginHttpRealIP:
		}

	}

	return pluginConfig, nil
}

func UpdatePluginConfigWhenDeleteCredential(pluginName string, credential providerDto.CredentialDto, config interface{}) ([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList, error) {
	// 只看全局配置对应的列表的第一个，因为当前(2023.02.23)只支持全局配置,且只会有一条配置记录
	pluginConfig, ok := config.([]*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList)
	if !ok {
		return nil, errors.Errorf("config is not type []*mseclient.GetPluginConfigResponseBodyDataGatewayConfigList")
	}

	currentConf := ""
	index := -1
	for idx, conf := range pluginConfig {
		// 只看全局配置对应的列表，因为当前(2023.02.23)只支持全局配置
		if *conf.ConfigLevel != 0 {
			continue
		}
		currentConf = *conf.Config
		index = idx
		break
	}

	if currentConf != "" {
		switch pluginName {
		case common.MsePluginKeyAuth:
			keyAutoConfig := mseDto.MsePluginKeyAuthConfig{}
			err := yaml.Unmarshal([]byte(currentConf), &keyAutoConfig)
			if err != nil {
				return nil, err
			}
			mapCredentialToConsumerName := updateWithDeleteCredential(credential, keyAutoConfig.Consumers, keyAutoConfig.Rules)

			keyAutoConfig.Consumers = make([]mseDto.Consumers, 0)
			for cred, consumer := range mapCredentialToConsumerName {
				keyAutoConfig.Consumers = append(keyAutoConfig.Consumers, mseDto.Consumers{
					Name:       consumer,
					Credential: cred,
				})
			}

			configBytes, err := yaml.Marshal(&keyAutoConfig)
			if err != nil {
				return nil, err
			}

			currentConf = string(configBytes)
			logrus.Infof("Yaml file content: \n%s\n", string(configBytes))
			pluginConfig[index].Config = &currentConf

		case common.MsePluginBasicAuth:
		case common.MsePluginHmacAuth:
		case common.MsePluginCustomResponse:
		case common.MsePluginRequestBlock:
		case common.MsePluginBotDetect:
		case common.MsePluginKeyRateLimit:
		case common.MsePluginHttp2Misdirect:
		case common.MsePluginJwtAuth:
		case common.MsePluginHttpRealIP:
		}

	}

	return pluginConfig, nil
}

func updateWithDeleteConsumer(consumerName string, consumers []mseDto.Consumers, rules []mseDto.Rules) (map[string]string, map[string][]string) {
	mapCredentialToConsumerName := make(map[string]string)
	for _, consumer := range consumers {
		if consumer.Name == consumerName {
			continue
		}
		mapCredentialToConsumerName[consumer.Credential] = consumer.Name
	}

	mapConsumerNameToRoutes := make(map[string][]string)
	for _, rule := range rules {
		for _, allow := range rule.Allow {
			if allow == consumerName {
				continue
			}
			if _, ok := mapConsumerNameToRoutes[allow]; !ok {
				mapConsumerNameToRoutes[allow] = make([]string, 0)
			}
			mapConsumerNameToRoutes[allow] = append(mapConsumerNameToRoutes[allow], rule.MatchRoute...)
		}
	}
	return mapCredentialToConsumerName, mapConsumerNameToRoutes
}

func updateWithDeleteCredential(credential providerDto.CredentialDto, consumers []mseDto.Consumers, rules []mseDto.Rules) map[string]string {
	mapCredentialToConsumerName := make(map[string]string)
	for _, consumer := range consumers {
		if consumer.Credential == credential.Key {
			continue
		}
		mapCredentialToConsumerName[consumer.Credential] = consumer.Name
	}

	return mapCredentialToConsumerName
}

// isInList 判断 ele 是否在 list 中
func isInList(list []string, ele string) bool {
	if ele == "" {
		// 空，则相当于忽略
		return true
	}
	if len(list) == 0 {
		return false
	}

	listMap := make(map[string]interface{})
	for _, v := range list {
		listMap[v] = nil
	}

	if _, ok := listMap[ele]; ok {
		return true
	} else {
		return false
	}
}

// deleteSubListFromList 从全序列 list 中删除子序列 sub
func deleteSubListFromList(list, sub []string) []string {
	if len(sub) == 0 {
		return list
	}

	result := make([]string, 0)
	listMap := make(map[string]interface{})
	subListMap := make(map[string]interface{})

	for _, v := range list {
		listMap[v] = nil
	}
	for _, v := range sub {
		subListMap[v] = nil
	}

	for k := range subListMap {
		if _, ok := listMap[k]; ok {
			delete(listMap, k)
		}
	}

	for k := range listMap {
		result = append(result, k)
	}

	return result
}

func getCredentialsAndRoutesMaps(consumers []mseDto.Consumers, rules []mseDto.Rules) (map[string]string, map[string][]string) {
	mapCredentialToConsumer := make(map[string]string)
	mapRouteToConsumers := make(map[string][]string)

	for _, cs := range consumers {
		mapCredentialToConsumer[cs.Credential] = cs.Name
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

	return mapCredentialToConsumer, mapRouteToConsumers
}
