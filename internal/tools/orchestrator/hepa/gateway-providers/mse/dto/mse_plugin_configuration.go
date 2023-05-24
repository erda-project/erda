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

package dto

type MsePluginConfig struct {
	Consumers  []Consumers `json:"consumers,omitempty" yaml:"consumers,omitempty"`
	Keys       []string    `json:"keys,omitempty" yaml:"keys,omitempty"`
	InQuery    bool        `json:"in_query,omitempty" yaml:"in_query,omitempty"`       // [key-auth 参数] [in_query 和 in_header 至少一个为 true] 配置 true 时，网关会尝试从 URL 参数中解析 API Key, 默认 true
	InHeader   bool        `json:"in_header,omitempty" yaml:"in_header,omitempty"`     // [key-auth 参数] [in_query 和 in_header 至少一个为 true] 配置 true 时，网关会尝试从 HTTP 请求头中解析 API Key, 默认 true
	DateOffset int         `json:"date_offset,omitempty" yaml:"date_offset,omitempty"` // [para-sign-auth、hmac-auth 参数], 单位 秒, 默认 300s
	Rules      []Rules     `json:"_rules_,omitempty" yaml:"_rules_,omitempty"`
}

type Consumers struct {
	Name string `json:"name" yaml:"name"`

	// key-auth and basic-auth
	Credential string `json:"credential,omitempty" yaml:"credential,omitempty"`

	// hmac-auth
	Key    string `json:"key,omitempty" yaml:"key,omitempty"`
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty"` // for hmac-auth only

	// for jwt-auth
	Issuer           string   `json:"issuer,omitempty" yaml:"issuer,omitempty"`                         // for jwt-auth only
	Jwks             string   `json:"jwks,omitempty" yaml:"jwks,omitempty"`                             // for jwt-auth only
	FromParams       []string `json:"from_params,omitempty" yaml:"from_params,omitempty"`               // for jwt-auth only   default: ["access_token"]
	FromCookies      []string `json:"from_cookies,omitempty" yaml:"from_cookies,omitempty"`             // for jwt-auth only   default: -
	KeepToken        bool     `json:"keep_token,omitempty" yaml:"keep_token,omitempty"`                 // for jwt-auth only   default: true
	ClockSkewSeconds int      `json:"clock_skew_seconds,omitempty" yaml:"clock_skew_seconds,omitempty"` // for jwt-auth only   default: 60
	//ClaimsToHeaders  []Object `json:"claims_to_headers,omitempty" yaml:"claims_to_headers,omitempty"` // for jwt-auth only   default: -   对象结构是啥不明确
	//FromHeaders      []Object `json:"from_headers,omitempty" yaml:"from_headers,omitempty"`           // for jwt-auth only   default: {"name":"Authorization","value_prefix":"Bearer "}  对象结构是啥不明确

	/*
		// for oauth2
		OAuthName    string `json:"oauth_name,omitempty" yaml:"auth_name,omitempty"`
		ClientId     string `json:"client_id,omitempty" yaml:"client_id,omitempty"`
		ClientSecret string `json:"client_secret,omitempty" yaml:"client_secret,omitempty"`
		// oauth2
		RedirectUrl interface{} `json:"redirect_uri,omitempty" yaml:"redirect_uri,omitempty"`
		// v2
		RedirectUrls []string `json:"redirect_uris,omitempty" yaml:"redirect_uris,omitempty"`
	*/
}

type Rules struct {
	MatchRoute []string `json:"_match_route_,omitempty" yaml:"_match_route_,omitempty"` // 路由生效(与 域名生效 二选一)
	Allow      []string `json:"allow,omitempty" yaml:"allow,omitempty"`
	// 暂时只支持路由生效
	// MatchDomain []string `json:"_match_domain_,omitempty" yaml:"_match_domain_,omitempty"` // 域名生效 (与 路由生效 二选一)

	// 以下为自定义开发 MSE 插件使用的配置
	// erda-para-sign-auth
	Consumers            []Consumers `json:"consumers,omitempty" yaml:"consumers,omitempty"`
	RequestBodySizeLimit int         `json:"request_body_size_limit,omitempty" yaml:"request_body_size_limit,omitempty"` // [erda-para-sign-auth 参数], 单位 Byte, 默认 10MB
	DateOffset           int         `json:"date_offset,omitempty" yaml:"date_offset,omitempty"`                         // [erda-para-sign-auth], 单位 秒, 默认 300s

	// erda-ip
	IPSource  string `json:"ip_source,omitempty" yaml:"ip_source,omitempty"`
	IpAclType string `json:"ip_acl_type,omitempty" yaml:"ip_acl_type,omitempty"`
	// 白名单对应一定要设置，黑名单可以不设置
	IpAclList []string `json:"ip_acl_list,omitempty" yaml:"ip_acl_list,omitempty"`

	// erda-sbac
	AccessControlAPI string   `json:"access_control_api,omitempty" yaml:"access_control_api,omitempty"`
	HttpMethods      []string `json:"http_methods,omitempty" yaml:"http_methods,omitempty"`
	MatchPatterns    []string `json:"match_patterns,omitempty" yaml:"match_patterns,omitempty"`
	WithHeaders      []string `json:"with_headers,omitempty" yaml:"with_headers,omitempty"`
	WithCookie       bool     `json:"with_cookie,omitempty" yaml:"with_cookie,omitempty"`
}

type SortConsumers struct {
	Consumers []Consumers
	By        func(p, q *Consumers) bool
}

func (cs SortConsumers) Len() int { // 重写 Len() 方法
	return len(cs.Consumers)
}
func (cs SortConsumers) Swap(i, j int) { // 重写 Swap() 方法
	cs.Consumers[i], cs.Consumers[j] = cs.Consumers[j], cs.Consumers[i]
}
func (cs SortConsumers) Less(i, j int) bool { // 重写 Less() 方法
	return cs.By(&cs.Consumers[i], &cs.Consumers[j])
}

type SortRules struct {
	Rules []Rules
	By    func(p, q *Rules) bool
}

func (cs SortRules) Len() int { // 重写 Len() 方法
	return len(cs.Rules)
}
func (cs SortRules) Swap(i, j int) { // 重写 Swap() 方法
	cs.Rules[i], cs.Rules[j] = cs.Rules[j], cs.Rules[i]
}
func (cs SortRules) Less(i, j int) bool { // 重写 Less() 方法
	return cs.By(&cs.Rules[i], &cs.Rules[j])
}
