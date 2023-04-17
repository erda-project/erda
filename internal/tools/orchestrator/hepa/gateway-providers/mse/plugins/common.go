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

	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

const (
	DEFAULT_MSE_ROUTE_NAME          = "route-erda-default"
	DEFAULT_MSE_CONSUMER_NAME       = "consumer-erda-default"
	DEFAULT_MSE_CONSUMER_CREDENTIAL = "2bda943c-ba2b-11ec-ba07-00163e1250b5"
	DEFAULT_MSE_CONSUMER_KEY        = "2bda943c-ba2b-11ec-ba07-00163e1250b5"
	DEFAULT_MSE_CONSUMER_SECRET     = "2bda943c-ba2b-11ec-ba07-00163e1250b5"

	DEFAULT_MSE_KEY_AUTH_CONFIG       = "# 配置必须字段的校验，如下例所示，要求插件配置必须存在 \"consumers\"、\"_rules_\" 字段\nconsumers: \n- key: 2bda943c-ba2b-11ec-ba07-00163e1250b5\n  secret: 2bda943c-ba2b-11ec-ba07-00163e1250b5\n  name: consumer-erda-default\nkeys:\n  - appKey\n  - x-app-key\nin_query: true\nin_header: true\n# 使用 _rules_ 字段进行细粒度规则配置\n_rules_:\n# 按路由名称匹配生效\n- _match_route_:\n  - route-erda-default\n  allow:\n  - consumer-erda-default"
	DEFAULT_MSE_HMAC_AUTH_CONFIG      = "# 配置必须字段的校验，如下例所示，要求插件配置必须存在 \"consumers\"、\"_rules_\" 字段\nconsumers: \n- key: 2bda943c-ba2b-11ec-ba07-00163e1250b5\n  secret: 2bda943c-ba2b-11ec-ba07-00163e1250b5\n  name: consumer-erda-default\n# 使用 _rules_ 字段进行细粒度规则配置\n_rules_:\n# 按路由名称匹配生效\n- _match_route_:\n  - route-erda-default\n  allow:\n  - consumer-erda-default"
	DEFAULT_MSE_PARA_SIGN_AUTH_CONFIG = "# 配置必须字段的校验，如下例所示，要求插件配置必须存在 \"_rules_\" 字段\n_rules_:\n- _match_route_:\n  - route-erda-default\n  request_body_size_limit: 10485760\n  date_offset: 600\n  consumers:\n  - name: consumer-erda-default\n    key: 2bda943c-ba2b-11ec-ba07-00163e1250b5\n    secret: 2bda943c-ba2b-11ec-ba07-00163e1250b5"

	MSE_PLUGIN_REQUEST_BODY_SIZE_LIMIT = 33554432
	MSE_PLUGIN_REQUEST_DATE_OFFSET     = 300
)

const (
	// Plugin Config Level
	MsePluginConfigLevelGlobal string = "global"
	MsePluginConfigLevelDomain string = "domain"
	MsePluginConfigLevelRoute  string = "route"

	MsePluginConfigLevelGlobalNumber int32 = 0
	MsePluginConfigLevelDomainNumber int32 = 1
	MsePluginConfigLevelRouteNumber  int32 = 2
)

const (
	REQUEST_BODY_SIZE_LIMIT int = 32 * 1024 * 1024 //32MB
)

type KeySecretConsumer struct {
	Name   string
	Key    string
	Secret string
}

func CreatePluginConfig(req *PluginReqDto, confList map[string][]mseclient.GetPluginConfigResponseBodyDataGatewayConfigList) (string, int64, error) {
	var configId int64 = -1
	pluginConfig := mseDto.MsePluginConfig{}

	// 只看全局配置对应的列表的第一个，因为当前(2023.02.23)只支持全局配置,且只会有一条配置记录
	if globalConfig, ok := confList[MsePluginConfigLevelGlobal]; ok {
		if len(globalConfig) > 0 {
			configId = *globalConfig[0].Id
			err := yaml.Unmarshal([]byte(*globalConfig[0].Config), &pluginConfig)
			if err != nil {
				return "", -1, err
			}
		}
	}

	matchRoutes := make([]string, 0)
	cons, ok := req.Config["whitelist"]
	if !ok {
		return "", -1, errors.Errorf("no whitelist in PluginReqDto Config")
	}
	consumers, ok := cons.([]mseDto.Consumers)
	if !ok {
		return "", -1, errors.Errorf("PluginReqDto.Config[whitelist] is not Type []Consumers ")
	}

	allows := make([]string, 0)
	for idx, cs := range consumers {
		if cs.Name == DEFAULT_MSE_CONSUMER_NAME {
			matchRoutes = append(matchRoutes, DEFAULT_MSE_ROUTE_NAME)
			switch req.Name {
			case common.MsePluginKeyAuth:
				consumers[idx].Credential = DEFAULT_MSE_CONSUMER_CREDENTIAL
			case common.MsePluginHmacAuth, common.MsePluginParaSignAuth:
				consumers[idx].Key = DEFAULT_MSE_CONSUMER_KEY
				consumers[idx].Secret = DEFAULT_MSE_CONSUMER_SECRET
			}
		}
		allows = append(allows, cs.Name)
	}

	if req.MSERouteName != "" {
		matchRoutes = append(matchRoutes, req.MSERouteName)
	}

	updateConfig := mseDto.MsePluginConfig{
		Consumers: consumers,
		Rules: []mseDto.Rules{
			{
				MatchRoute: matchRoutes,
				Allow:      allows,
			},
		},
	}

	var err error = nil
	switch req.Name {
	case common.MsePluginKeyAuth:
		pluginConfig, err = mergeKeyAuthConfig(pluginConfig, updateConfig)
		if err != nil {
			return "", -1, err
		}

	case common.MsePluginHmacAuth:
		pluginConfig, err = mergeHmacAuthConfig(pluginConfig, updateConfig)
		if err != nil {
			return "", -1, err
		}
	case common.MsePluginParaSignAuth:
		updateConfig = mseDto.MsePluginConfig{
			Rules: []mseDto.Rules{
				{
					MatchRoute:           matchRoutes,
					Consumers:            consumers,
					RequestBodySizeLimit: MSE_PLUGIN_REQUEST_BODY_SIZE_LIMIT,
					//DateOffset:           MSE_PLUGIN_REQUEST_DATE_OFFSET,
				},
			},
		}
		pluginConfig, err = mergeParaSignAuthConfig(pluginConfig, updateConfig)
		if err != nil {
			return "", -1, err
		}
	}

	configBytes, _ := yaml.Marshal(&pluginConfig)
	logrus.Infof("merge %s config result:\n************************************************************\n%s\n********************************************************", req.Name, string(configBytes))

	return string(configBytes), configId, nil
}

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
		msePluginConfig := mseDto.MsePluginConfig{}
		err := yaml.Unmarshal([]byte(currentConf), &msePluginConfig)
		if err != nil {
			return nil, err
		}
		mapCredentialToConsumerName, mapKeyToConsumerName, mapKeyToConsumerSecret, mapConsumerNameToRoutes, customPluginRules := updateWithDeleteConsumer(pluginName, consumerName, msePluginConfig.Consumers, msePluginConfig.Rules)

		switch pluginName {
		case common.MsePluginKeyAuth:
			msePluginConfig.Consumers = make([]mseDto.Consumers, 0)
			for cred, consumer := range mapCredentialToConsumerName {
				msePluginConfig.Consumers = append(msePluginConfig.Consumers, mseDto.Consumers{
					Name:       consumer,
					Credential: cred,
				})
			}
			msePluginConfig.Rules = make([]mseDto.Rules, 0)
			for allow, routes := range mapConsumerNameToRoutes {
				msePluginConfig.Rules = append(msePluginConfig.Rules, mseDto.Rules{
					MatchRoute: routes,
					Allow:      []string{allow},
				})
			}

		case common.MsePluginBasicAuth:
		case common.MsePluginHmacAuth:
			msePluginConfig.Consumers = make([]mseDto.Consumers, 0)
			for key, consumer := range mapKeyToConsumerName {
				msePluginConfig.Consumers = append(msePluginConfig.Consumers, mseDto.Consumers{
					Name:   consumer,
					Key:    key,
					Secret: mapKeyToConsumerSecret[key],
				})
			}
			msePluginConfig.Rules = make([]mseDto.Rules, 0)
			for allow, routes := range mapConsumerNameToRoutes {
				msePluginConfig.Rules = append(msePluginConfig.Rules, mseDto.Rules{
					MatchRoute: routes,
					Allow:      []string{allow},
				})
			}

		case common.MsePluginParaSignAuth:
			msePluginConfig.Rules = customPluginRules
		case common.MsePluginCustomResponse:
		case common.MsePluginRequestBlock:
		case common.MsePluginBotDetect:
		case common.MsePluginKeyRateLimit:
		case common.MsePluginHttp2Misdirect:
		case common.MsePluginJwtAuth:
		case common.MsePluginHttpRealIP:
		}
		configBytes, err := yaml.Marshal(&msePluginConfig)
		if err != nil {
			return nil, err
		}

		currentConf = string(configBytes)
		logrus.Debugf("plugin %s config Yaml file content: \n%s\n", pluginName, string(configBytes))
		pluginConfig[index].Config = &currentConf
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
		msePluginConfig := mseDto.MsePluginConfig{}
		err := yaml.Unmarshal([]byte(currentConf), &config)
		if err != nil {
			return nil, err
		}
		mapCredentialToConsumerName, mapKeyToConsumerName, mapKeyToConsumerSecret, customPluginConfig := updateWithDeleteCredential(pluginName, credential, msePluginConfig)

		switch pluginName {
		case common.MsePluginKeyAuth:
			msePluginConfig.Consumers = make([]mseDto.Consumers, 0)
			for cred, consumer := range mapCredentialToConsumerName {
				msePluginConfig.Consumers = append(msePluginConfig.Consumers, mseDto.Consumers{
					Name:       consumer,
					Credential: cred,
				})
			}
		case common.MsePluginBasicAuth:
		case common.MsePluginHmacAuth:
			msePluginConfig.Consumers = make([]mseDto.Consumers, 0)
			for key, consumer := range mapKeyToConsumerName {
				msePluginConfig.Consumers = append(msePluginConfig.Consumers, mseDto.Consumers{
					Name:   consumer,
					Key:    key,
					Secret: mapKeyToConsumerSecret[key],
				})
			}

		case common.MsePluginParaSignAuth:
			msePluginConfig = customPluginConfig
		case common.MsePluginCustomResponse:
		case common.MsePluginRequestBlock:
		case common.MsePluginBotDetect:
		case common.MsePluginKeyRateLimit:
		case common.MsePluginHttp2Misdirect:
		case common.MsePluginJwtAuth:
		case common.MsePluginHttpRealIP:
		}

		configBytes, err := yaml.Marshal(&msePluginConfig)
		if err != nil {
			return nil, err
		}

		currentConf = string(configBytes)
		logrus.Debugf("Yaml file content for plugin %s: \n%s\n", pluginName, string(configBytes))
		pluginConfig[index].Config = &currentConf
	}

	return pluginConfig, nil
}

func updateWithDeleteConsumer(pluginName, consumerName string, consumers []mseDto.Consumers, rules []mseDto.Rules) (map[string]string, map[string]string, map[string]string, map[string][]string, []mseDto.Rules) {
	mapCredentialToConsumerName := make(map[string]string)
	mapKeyToConsumerName := make(map[string]string)
	mapKeyToConsumerSecret := make(map[string]string)
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

	for _, consumer := range consumers {
		if consumer.Name == consumerName {
			continue
		}
		switch pluginName {
		case common.MsePluginKeyAuth:
			mapCredentialToConsumerName[consumer.Credential] = consumer.Name
		case common.MsePluginHmacAuth:
			mapKeyToConsumerName[consumer.Key] = consumer.Name
			mapKeyToConsumerSecret[consumer.Key] = consumer.Secret
		}
	}

	switch pluginName {
	case common.MsePluginKeyAuth:
		return mapCredentialToConsumerName, nil, nil, mapConsumerNameToRoutes, nil
	case common.MsePluginHmacAuth:
		return nil, mapKeyToConsumerName, mapKeyToConsumerSecret, mapConsumerNameToRoutes, nil
	case common.MsePluginParaSignAuth:
		paraSignConfigRules := make([]mseDto.Rules, 0)
		for _, rule := range rules {
			ruleNewConsumers := make([]mseDto.Consumers, 0)
			for _, consumer := range rule.Consumers {
				if consumer.Name == consumerName {
					continue
				}
				ruleNewConsumers = append(ruleNewConsumers, consumer)
			}
			rule.Consumers = ruleNewConsumers
			if len(ruleNewConsumers) > 0 {
				rule.Consumers = ruleNewConsumers
				paraSignConfigRules = append(paraSignConfigRules, rule)
			}
		}
		return nil, nil, nil, nil, paraSignConfigRules
	default:
		return nil, nil, nil, nil, nil
	}
}

func updateWithDeleteCredential(pluginName string, credential providerDto.CredentialDto, msePluginConfig mseDto.MsePluginConfig) (map[string]string, map[string]string, map[string]string, mseDto.MsePluginConfig) {
	mapCredentialToConsumerName := make(map[string]string)
	mapKeyToConsumerName := make(map[string]string)
	mapKeyToConsumerSecret := make(map[string]string)
	var customPluginConfig mseDto.MsePluginConfig

	for _, consumer := range msePluginConfig.Consumers {
		if consumer.Credential == credential.Key {
			continue
		}
		switch pluginName {
		case common.MsePluginKeyAuth:
			mapCredentialToConsumerName[consumer.Credential] = consumer.Name
		case common.MsePluginHmacAuth:
			mapKeyToConsumerName[consumer.Key] = consumer.Name
			mapKeyToConsumerSecret[consumer.Key] = consumer.Secret
		case common.MsePluginParaSignAuth:
			mapKeyToConsumerName[consumer.Key] = consumer.Name
			mapKeyToConsumerSecret[consumer.Key] = consumer.Secret
		}
	}

	if pluginName == common.MsePluginParaSignAuth {
		for _, rule := range msePluginConfig.Rules {
			consumers := make([]mseDto.Consumers, 0)
			for _, consumer := range rule.Consumers {
				if consumer.Credential != credential.Key {
					consumers = append(consumers, consumer)
				}
			}
			if len(consumers) > 0 {
				rule.Consumers = consumers
				customPluginConfig.Rules = append(customPluginConfig.Rules, rule)
			}
		}
	}

	return mapCredentialToConsumerName, mapKeyToConsumerName, mapKeyToConsumerSecret, customPluginConfig
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
