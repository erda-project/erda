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
)

var MapClusterNameToMSEPluginNameToPluginID map[string]map[string]*int64

func init() {
	MapClusterNameToMSEPluginNameToPluginID = make(map[string]map[string]*int64)
}
