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

package common

const (
	MseVersion                            = "mse-1.2.18"
	MseProviderName                       = "MSE"
	MseDefaultServerEndpoint              = "mse.cn-hangzhou.aliyuncs.com"
	MseBurstMultiplier1X                  = "1"
	MseBurstMultiplier2X                  = "2"
	MseBurstMultiplier3X                  = "3"
	MseBurstMultiplier4X                  = "4"
	MseIngressControllerAckNamespace      = "mse-ingress-controller"
	MseIngressControllerAckDeploymentName = "ack-mse-ingress-controller"
	MseNeedDropAnnotation                 = "need_drop_annotation"
)

// MSE 支持的插件名称及ID (通过 MSE 的获取网关插件列表的 API 获取，并非随意自定义)
const (
	//Plugin name
	MsePluginKeyAuth         string = "key-auth"
	MsePluginBasicAuth       string = "basic-auth"
	MsePluginHmacAuth        string = "hmac-auth"
	MsePluginCustomResponse  string = "custom-response"
	MsePluginRequestBlock    string = "request-block"
	MsePluginBotDetect       string = "bot-detect"
	MsePluginKeyRateLimit    string = "key-rate-limit"
	MsePluginHttp2Misdirect  string = "http2-misdirect"
	MsePluginJwtAuth         string = "jwt-auth"
	MsePluginHttpRealIP      string = "http-real-ip"
	MsePluginEDASServiceAuth string = "edas-service-auth"
	MsePluginWaf             string = "waf"
	MsePluginParaSignAuth    string = "erda-para-sign-auth"
	MsePluginIP              string = "erda-ip"
	MsePluginSbac            string = "erda-sbac"
	MsePluginCsrf            string = "erda-csrf"
)

// MSE 自定义插件使用到的一些常量名称
const (
	MseErdaIpRouteSwitch         = "MSE_ERDA_IP_ROUTE_SWITCH"
	MseErdaIpIpSource            = "IP_SOURCE"
	MseErdaIpAclType             = "ACL_TYPE"
	MseErdaIpAclList             = "ACL_LIST"
	MseErdaIpAclBlack            = "black"
	MseErdaIpAclWhite            = "white"
	MseErdaIpSourceRemoteIP      = "remote-ip"
	MseErdaIpSourceXRealIP       = "x-real-ip"
	MseErdaIpSourceXForwardedFor = "x-forwarded-for"

	MseErdaSBACRouteSwitch               = "MSE_ERDA_SBAC_ROUTE_SWITCH"
	MseErdaSBACAccessControlAPI          = "http://test-sbac.default.svc.cluster.local:8080/"
	MseErdaSBACConfigAccessControlAPI    = "access_control_api"
	MseErdaSBACConfigMatchPatterns       = "patterns"
	MseErdaSBACConfigHttpMethods         = "methods"
	MseErdaSBACConfigWithHeaders         = "with_headers"
	MseErdaSBACConfigWithCookie          = "with_cookie"
	MseErdaSBACConfigDefaultMatchPattern = "^/"
	MseErdaSBACConfigDefaultWithHeader   = "*"

	MseErdaCSRFRouteSwitch          = "MSE_ERDA_CSRF_ROUTE_SWITCH"
	MseErdaCSRFConfigUserCookie     = "biz_cookie"
	MseErdaCSRFConfigExcludedMethod = "excluded_method"
	MseErdaCSRFConfigTokenKey       = "token_key"
	MseErdaCSRFConfigTokenCookie    = "token_cookie"
	MseErdaCSRFConfigTokenDomain    = "biz_domain"
	MseErdaCSRFConfigCookieSecure   = "secure_cookie"
	MseErdaCSRFConfigValidTTL       = "valid_ttl"
	MseErdaCSRFConfigRefreshTTL     = "refresh_ttl"
	MseErdaCSRFConfigErrStatus      = "err_status"
	MseErdaCSRFConfigErrMsg         = "err_message"
	MseErdaCSRFConfigSecret         = "jwt_secret"

	MseErdaCSRFDefaultUserCookie   = "uc-token"
	MseErdaCSRFDefaultTokenName    = "csrf-token"
	MseErdaCSRFDefaultTokenDomain  = ""
	MseErdaCSRFDefaultCookieSecure = false
	MseErdaCSRFDefaultValidTTL     = int64(1800)
	MseErdaCSRFDefaultRefreshTTL   = int64(10)
	MseErdaCSRFDefaultErrStatus    = int64(403)
	MseErdaCSRFDefaultErrMsg       = `{"message":"This form has expired. Please refresh and try again."}`
	MseErdaCSRFDefaultJWTSecret    = "e796dce47e561ff926d2916144b8e4bf"
)

var MapClusterNameToMSEPluginNameToPluginID map[string]map[string]*int64

func init() {
	MapClusterNameToMSEPluginNameToPluginID = make(map[string]map[string]*int64)
}
