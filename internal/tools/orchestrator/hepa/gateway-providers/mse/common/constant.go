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
	Mse_Version                           = "mse-1.2.15"
	Mse_Provider_Name                     = "MSE"
	MseDefaultServerEndpoint              = "mse.cn-hangzhou.aliyuncs.com"
	MseBurstMultiplier                    = "2"
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
	MsePluginParaSignAuth    string = "para-sign-auth"

	// Plugin ID
	MsePluginKeyAuthID         int64 = 1
	MsePluginBasicAuthID       int64 = 2
	MsePluginHmacAuthID        int64 = 3
	MsePluginCustomResponseID  int64 = 4
	MsePluginRequestBlockID    int64 = 5
	MsePluginBotDetectID       int64 = 6
	MsePluginKeyRateLimitID    int64 = 7
	MsePluginHttp2MisdirectID  int64 = 23
	MsePluginJwtAuthID         int64 = 34
	MsePluginHttpRealIPID      int64 = 43
	MsePluginEDASServiceAuthID int64 = 114
	MsePluginWafID             int64 = 119
	MsePluginParaSignAuthID    int64 = 129
)

var MSEPluginNameToID map[string]*int64

func init() {
	MSEPluginNameToID = make(map[string]*int64)
	keyAuthID := MsePluginKeyAuthID
	MSEPluginNameToID[MsePluginKeyAuth] = &keyAuthID

	basicAuthID := MsePluginBasicAuthID
	MSEPluginNameToID[MsePluginBasicAuth] = &basicAuthID

	hmacAuthID := MsePluginHmacAuthID
	MSEPluginNameToID[MsePluginHmacAuth] = &hmacAuthID

	customResponseID := MsePluginCustomResponseID
	MSEPluginNameToID[MsePluginCustomResponse] = &customResponseID

	requestBlockID := MsePluginRequestBlockID
	MSEPluginNameToID[MsePluginRequestBlock] = &requestBlockID

	botDetectID := MsePluginBotDetectID
	MSEPluginNameToID[MsePluginBotDetect] = &botDetectID

	keyRateLimitID := MsePluginKeyRateLimitID
	MSEPluginNameToID[MsePluginKeyRateLimit] = &keyRateLimitID

	http2MisdirectID := MsePluginHttp2MisdirectID
	MSEPluginNameToID[MsePluginHttp2Misdirect] = &http2MisdirectID

	jwtAuthID := MsePluginJwtAuthID
	MSEPluginNameToID[MsePluginJwtAuth] = &jwtAuthID

	httpRealIpID := MsePluginHttpRealIPID
	MSEPluginNameToID[MsePluginHttpRealIP] = &httpRealIpID
}
