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

type ErdaCSRFConfig struct {
	MatchRoute     string   `json:"_match_route_,omitempty" yaml:"_match_route_,omitempty"`
	UserCookie     string   `json:"userCookie" yaml:"userCookie"`         // 鉴别 Cookie
	ExcludedMethod []string `json:"excludedMethod" yaml:"excludedMethod"` //关闭校验的 HTTP Methods 列表
	TokenName      string   `json:"tokenName" yaml:"tokenName"`           // token 名称
	TokenDomain    string   `json:"tokenDomain" yaml:"tokenDomain"`       // token 域名
	CookieSecure   bool     `json:"cookieSecure" yaml:"cookieSecure"`     // secure 开关
	ValidTTL       int64    `json:"validTTL" yaml:"validTTL"`             // token 过期时间(单位s)
	RefreshTTL     int64    `json:"refreshTTL" yaml:"refreshTTL"`         // token 更新周期(单位s)
	ErrStatus      int64    `json:"errStatus" yaml:"errStatus"`           // 失败状态码
	ErrMsg         string   `json:"errMsg" yaml:"errMsg"`                 // 失败应答
	JWTSecret      string   `json:"jwtSecret" yaml:"jwtSecret"`           // 用于加密的 Secret
	Disable        bool
}

/* erda-csrf 插件配置格式示例
_rules_:
- _match_route_:
  - route-erda-default
  userCookie: "uc-token"
  tokenName: "csrf-token"
  tokenDomain: ""
  errMsg: "{\"message\":\"This form has expired. Please refresh and try again.\"}"
  cookieSecure: false
  validTTL: 1800
  refreshTTL: 300
  errStatus: 403
  jwtSecret: "e796dce47e561ff926d2916144b8e4bf"
  excludedMethod:
  - PATCH
  - DELETE
*/

func mergeErdaCSRFConfig(currentErdaCSRFConfig, updateErdaCSRFConfig mseDto.MsePluginConfig, updateForDisable bool) (mseDto.MsePluginConfig, error) {
	configBytes, _ := yaml.Marshal(&currentErdaCSRFConfig)
	logrus.Debugf("Current ErdaCSRF config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))
	configBytes, _ = yaml.Marshal(&updateErdaCSRFConfig)
	logrus.Debugf("Update ErdaCSRF config result Yaml file content Breore merge:\n**********************************************************\n%s\n*******************************************", string(configBytes))

	// erda-csrf 的 路由唯一性
	// 当前配置项转换
	mapCurrentErdaCSRFConfig := getErdaCSRFRouteNameToMatchRouteMap(currentErdaCSRFConfig, true, updateForDisable)
	logrus.Debugf("mapCurrentErdaCSRFConfig=%+v", mapCurrentErdaCSRFConfig)

	// 更新配置项转换
	mapUpdateErdaCSRFConfig := getErdaCSRFRouteNameToMatchRouteMap(updateErdaCSRFConfig, false, updateForDisable)
	logrus.Debugf("mapUpdateErdaCSRFConfig=%+v", mapUpdateErdaCSRFConfig)

	// 更新 routes
	for route, erdaCSRFConfig := range mapUpdateErdaCSRFConfig {
		if _, ok := mapCurrentErdaCSRFConfig[route]; ok {
			// case 1: 如果存在,则判断当前更新是否是删除，是则删除，不是则直接更新
			if erdaCSRFConfig.Disable {
				delete(mapCurrentErdaCSRFConfig, route)
			} else {
				mapCurrentErdaCSRFConfig[route] = erdaCSRFConfig
			}
		} else {
			// case 2: 如果不存在，则当前更新时新增，则加入
			if !erdaCSRFConfig.Disable {
				mapCurrentErdaCSRFConfig[route] = erdaCSRFConfig
			}
		}
	}

	rules := make([]mseDto.Rules, 0)
	if len(mapCurrentErdaCSRFConfig) == 0 {
		mapCurrentErdaCSRFConfig[MseDefaultRouteName] = ErdaCSRFConfig{
			MatchRoute:     MseDefaultRouteName,
			UserCookie:     common.MseErdaCSRFDefaultUserCookie,
			ExcludedMethod: []string{"GET", "HEAD", "OPTIONS", "TRACE"},
			TokenName:      common.MseErdaCSRFDefaultTokenName,
			TokenDomain:    common.MseErdaCSRFDefaultTokenDomain,
			CookieSecure:   common.MseErdaCSRFDefaultCookieSecure,
			ValidTTL:       common.MseErdaCSRFDefaultValidTTL,
			RefreshTTL:     common.MseErdaCSRFDefaultRefreshTTL,
			ErrStatus:      common.MseErdaCSRFDefaultErrStatus,
			ErrMsg:         common.MseErdaCSRFDefaultErrMsg,
			JWTSecret:      common.MseErdaCSRFDefaultJWTSecret,
			Disable:        false,
		}
	}
	for route, erdaCSRFConfig := range mapCurrentErdaCSRFConfig {
		rule := mseDto.Rules{
			MatchRoute:     []string{route},
			UserCookie:     erdaCSRFConfig.UserCookie,
			ExcludedMethod: erdaCSRFConfig.ExcludedMethod,
			TokenName:      erdaCSRFConfig.TokenName,
			TokenDomain:    erdaCSRFConfig.TokenDomain,
			CookieSecure:   erdaCSRFConfig.CookieSecure,
			ValidTTL:       erdaCSRFConfig.ValidTTL,
			RefreshTTL:     erdaCSRFConfig.RefreshTTL,
			ErrStatus:      erdaCSRFConfig.ErrStatus,
			ErrMsg:         erdaCSRFConfig.ErrMsg,
			JWTSecret:      erdaCSRFConfig.JWTSecret,
		}

		if len(rule.ExcludedMethod) == 0 {
			rule.ExcludedMethod = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
		}

		rules = append(rules, rule)
	}

	result := mseDto.MsePluginConfig{
		Rules: rules,
	}

	return result, nil
}

func getErdaCSRFRouteNameToMatchRouteMap(pluginConfig mseDto.MsePluginConfig, isCurrentConfig bool, isUpdateForDelete bool) map[string]ErdaCSRFConfig {
	routeNameToConfig := make(map[string]ErdaCSRFConfig)

	for _, rule := range pluginConfig.Rules {
		for _, route := range rule.MatchRoute {
			erdaCSRFConfig := ErdaCSRFConfig{
				MatchRoute:     route,
				UserCookie:     rule.UserCookie,
				ExcludedMethod: rule.ExcludedMethod,
				TokenName:      rule.TokenName,
				TokenDomain:    rule.TokenDomain,
				CookieSecure:   rule.CookieSecure,
				ValidTTL:       rule.ValidTTL,
				RefreshTTL:     rule.RefreshTTL,
				ErrStatus:      rule.ErrStatus,
				ErrMsg:         rule.ErrMsg,
				JWTSecret:      rule.JWTSecret,
			}
			if len(rule.ExcludedMethod) == 0 {
				erdaCSRFConfig.ExcludedMethod = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
			}
			if isCurrentConfig {
				erdaCSRFConfig.Disable = false
			} else {
				erdaCSRFConfig.Disable = isUpdateForDelete
			}
			routeNameToConfig[route] = erdaCSRFConfig
		}
	}

	return routeNameToConfig
}

func getErdaCSRFSourceConfig(config map[string]interface{}) (excludedMethod []string, userCookie string, tokenName, tokenDomain, errMsg, jSecret string, validTTL, refreshTTL, errStatus int64, cookieSecure bool, disable bool, err error) {
	userCookiesArray, ok := config[common.MseErdaCSRFConfigUserCookie]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set useCookie for plugin %s", common.MsePluginCsrf)
	}
	userCookies, ok := userCookiesArray.([]string)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("useCookie for plugin %s is invalid []string", common.MsePluginCsrf)
	}
	userCookie = userCookies[0]

	excludedMethod = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	methods, ok := config[common.MseErdaCSRFConfigExcludedMethod]
	if ok {
		excludedMethod, ok = methods.([]string)
		if !ok {
			return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("excludedMethod for plugin %s is invalid []string", common.MsePluginCsrf)
		}
	}

	tn, ok := config[common.MseErdaCSRFConfigTokenCookie]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set tokenName for plugin %s", common.MsePluginCsrf)
	}
	tokenName, ok = tn.(string)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("tokenName for plugin %s is invalid string", common.MsePluginCsrf)
	}

	tokenDomain = ""
	tnDomain, ok := config[common.MseErdaCSRFConfigTokenDomain]
	if ok {
		tokenDomain, ok = tnDomain.(string)
		if !ok {
			return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("tokenDomain for plugin %s is invalid string", common.MsePluginCsrf)
		}
	}

	// 不存在表示设置为 false
	cookieSecure = false
	cookie, ok := config[common.MseErdaCSRFConfigCookieSecure]
	if ok {
		cookieSecure, ok = cookie.(bool)
		if !ok {
			return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("cookieSecure for plugin %s is invalid bool", common.MsePluginCsrf)
		}
	}

	vttl, ok := config[common.MseErdaCSRFConfigValidTTL]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set validTTL for plugin %s", common.MsePluginCsrf)
	}
	validTTL, ok = vttl.(int64)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("validTTL for plugin %s is invalid int64", common.MsePluginCsrf)
	}

	rttl, ok := config[common.MseErdaCSRFConfigRefreshTTL]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set refreshTTL for plugin %s", common.MsePluginCsrf)
	}
	refreshTTL, ok = rttl.(int64)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("refreshTTL for plugin %s is invalid int64", common.MsePluginCsrf)
	}

	eStatus, ok := config[common.MseErdaCSRFConfigErrStatus]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set errStatus for plugin %s", common.MsePluginCsrf)
	}
	errStatus, ok = eStatus.(int64)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("errStatus for plugin %s is invalid int64", common.MsePluginCsrf)
	}

	eMSg, ok := config[common.MseErdaCSRFConfigErrMsg]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set errMsg for plugin %s", common.MsePluginCsrf)
	}
	errMsg, ok = eMSg.(string)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("errMsg for plugin %s is invalid string", common.MsePluginCsrf)
	}

	js, ok := config[common.MseErdaCSRFConfigSecret]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set jwt_secret for plugin %s", common.MsePluginCsrf)
	}
	jSecret, ok = js.(string)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("jwt_secret for plugin %s is invalid string", common.MsePluginCsrf)
	}

	routeSwitch, ok := config[common.MseErdaCSRFRouteSwitch]
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("not set %s for plugin %s", common.MseErdaCSRFRouteSwitch, common.MsePluginCsrf)
	}
	enable, ok := routeSwitch.(bool)
	if !ok {
		return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, false, errors.Errorf("%s for plugin %s is invalid bool", common.MseErdaCSRFRouteSwitch, common.MsePluginCsrf)
	}

	disable = !enable

	return excludedMethod, userCookie, tokenName, tokenDomain, errMsg, jSecret, validTTL, refreshTTL, errStatus, cookieSecure, disable, nil
}
