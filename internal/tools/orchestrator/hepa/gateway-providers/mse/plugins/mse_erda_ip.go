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

	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/common"
	mseDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/mse/dto"
)

type ErdaIPConfig struct {
	MatchRoute string `json:"_match_route_,omitempty" yaml:"_match_route_,omitempty"`
	IPSource   string `json:"ip_source,omitempty" yaml:"ip_source,omitempty"`
	IpAclType  string `json:"ip_acl_type,omitempty" yaml:"ip_acl_type,omitempty"`
	// 白名单对应一定要设置，黑名单可以不设置
	IpAclList []string `json:"ip_acl_list,omitempty" yaml:"ip_acl_list,omitempty"`
	Disable   bool
}

var MSE_ERDA_IP_DEFALUT_ACL_LIST = []string{"10.10.10.10", "11.12.13.0/24"}

/* erda-ip 插件配置格式示例
_rules_:
- _match_route_:
  - route-erda-default
  ip_source: "x-forwarded-for"
  ip_acl_type: "black"
  ip_acl_list:
  - 10.10.10.10
  - 10.12.13.0/24
*/

func mergeErdaIPConfig(currentParaSignAuthConfig, updateParaSignAuthConfig mseDto.MsePluginConfig, updateForDisable bool) (mseDto.MsePluginConfig, error) {
	configBytes, _ := yaml.Marshal(&currentParaSignAuthConfig)
	logrus.Debugf("Current ErdaIP config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))
	configBytes, _ = yaml.Marshal(&updateParaSignAuthConfig)
	logrus.Debugf("Update ErdaIP config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))

	// erda-ip 的 路由唯一性
	// 当前配置项转换
	mapCurrentErdaIPConfig := getErdaIPRouteNameToMatchRouteMap(currentParaSignAuthConfig, true, updateForDisable)
	logrus.Debugf("mapCurrentErdaIPConfig=%+v", mapCurrentErdaIPConfig)

	// 更新配置项转换
	mapUpdateErdaIPConfig := getErdaIPRouteNameToMatchRouteMap(updateParaSignAuthConfig, false, updateForDisable)
	logrus.Debugf("mapUpdateErdaIPConfig=%+v", mapUpdateErdaIPConfig)

	// 更新 routes
	for route, erdaIPConfig := range mapUpdateErdaIPConfig {
		if _, ok := mapCurrentErdaIPConfig[route]; ok {
			// case 1: 如果存在,则判断当前更新是否是删除，是则删除，不是则直接更新
			if erdaIPConfig.Disable {
				delete(mapCurrentErdaIPConfig, route)
			} else {
				mapCurrentErdaIPConfig[route] = erdaIPConfig
			}
		} else {
			// case 2: 如果不存在，则当前更新时新增，则加入
			if !erdaIPConfig.Disable {
				mapCurrentErdaIPConfig[route] = erdaIPConfig
			}
		}
	}

	rules := make([]mseDto.Rules, 0)
	if len(mapCurrentErdaIPConfig) == 0 {
		mapCurrentErdaIPConfig[MseDefaultRouteName] = ErdaIPConfig{
			MatchRoute: "DEFAULT_MSE_ROUTE_NAME",
			IPSource:   common.MseErdaIpSourceXForwardedFor,
			IpAclType:  common.MseErdaIpAclWhite,
			IpAclList:  MSE_ERDA_IP_DEFALUT_ACL_LIST,
			Disable:    false,
		}
	}
	for route, erdaIPConfig := range mapCurrentErdaIPConfig {
		rule := mseDto.Rules{
			MatchRoute: []string{route},
			IPSource:   erdaIPConfig.IPSource,
			IpAclType:  erdaIPConfig.IpAclType,
			IpAclList:  erdaIPConfig.IpAclList,
		}
		rules = append(rules, rule)
	}

	result := mseDto.MsePluginConfig{
		Rules: rules,
	}

	return result, nil
}

func getErdaIPRouteNameToMatchRouteMap(pluginConfig mseDto.MsePluginConfig, isCurrentConfig bool, isUpdateForDelete bool) map[string]ErdaIPConfig {
	routeNameToConfig := make(map[string]ErdaIPConfig)

	for _, rule := range pluginConfig.Rules {
		for _, route := range rule.MatchRoute {
			erdaIPConfig := ErdaIPConfig{
				MatchRoute: route,
				IPSource:   rule.IPSource,
				IpAclType:  rule.IpAclType,
				IpAclList:  rule.IpAclList,
			}
			if isCurrentConfig {
				erdaIPConfig.Disable = false
			} else {
				erdaIPConfig.Disable = isUpdateForDelete
			}
			routeNameToConfig[route] = erdaIPConfig
		}
	}

	return routeNameToConfig
}

func getErdaIPSourceConfig(config map[string]interface{}) (ipSource string, ipAclType string, ipAclList []string, disable bool, err error) {

	source, ok := config[common.MseErdaIpIpSource]
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("not set ip_source for plugin %s", common.MsePluginIP)
	}
	ipSource, ok = source.(string)
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("ip_source for plugin %s is invalid string", common.MsePluginIP)
	}

	aclType, ok := config[common.MseErdaIpAclType]
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("not set ip_acl_type for plugin %s", common.MsePluginIP)
	}
	ipAclType, ok = aclType.(string)
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("ip_acl_type for plugin %s is invalid string", common.MsePluginIP)
	}

	aclList, ok := config[common.MseErdaIpAclList]
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("not set ip_acl_list for plugin %s", common.MsePluginIP)
	}
	ipAclList, ok = aclList.([]string)
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("ip_acl_list for plugin %s is invalid []string", common.MsePluginIP)
	}
	if len(ipAclList) == 0 && ipAclType == common.MseErdaIpAclWhite {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("can not set ip_acl_list nil for plugin %s with ip_acl_type %s", common.MsePluginIP, common.MseErdaIpAclWhite)
	}

	routeSwitch, ok := config[common.MseErdaIpRouteSwitch]
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("not set ip_acl_type for plugin %s", common.MsePluginIP)
	}
	enable, ok := routeSwitch.(bool)
	if !ok {
		return ipSource, ipAclType, ipAclList, false, errors.Errorf("ip_acl_type for plugin %s is invalid string", common.MsePluginIP)
	}

	disable = !enable

	return ipSource, ipAclType, ipAclList, disable, nil
}
