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

package orm

// 网关 Credential 信息，用于 MSE Gateway 网关场景，存储调用方授权信息
type GatewayCredential struct {
	ConsumerId   string `json:"consumer_id" xorm:"not null default '' comment('消费者id') unique VARCHAR(128)"`
	ConsumerName string `json:"consumer_name" xorm:"not null default '' comment('消费者名称') VARCHAR(128)"`
	PluginName   string `json:"plugin_name" xorm:"not null default '' comment('插件名称') VARCHAR(128)"`

	OrgName   string `json:"org_name" xorm:"not null default '' comment('组织 名称') VARCHAR(50)"`
	OrgId     string `json:"org_id" xorm:"not null comment('组织 ID') VARCHAR(32)"`
	ProjectId string `json:"project_id" xorm:"not null comment('项目 ID') VARCHAR(32)"`
	Env       string `json:"env" xorm:"not null comment('部署环境') VARCHAR(32)"`
	Az        string `json:"az" xorm:"not null comment('集群名称') VARCHAR(32)"`

	// key-auth and basic-auth
	//Credential string `json:"credential"  xorm:"not null default '' comment('插件名称') VARCHAR(128)"`

	// key-auth、basic-auth and hmac-auth
	Key string `json:"key"  xorm:"auth_key not null default '' comment('hmac-auth 对应的 key') VARCHAR(128)"`

	// sign-auth, hmac-auth
	Secret string `json:"secret" xorm:"not null default '' comment('hmac-auth 对应的 secret') VARCHAR(128)"` // for hmac-auth only
	// hmac-auth
	Username string `json:"username" xorm:"not null default '' comment('用户名称') unique VARCHAR(128)"`

	// JWT-AUTH
	Issuer           string `json:"issuer" xorm:"not null default '' comment('jwt-auth 对应的 issuer') VARCHAR(128)"`                // for jwt-auth only
	Jwks             string `json:"jwks" xorm:"not null default '' comment('jwt-auth 对应的 jwks') VARCHAR(4096)"`                   // for jwt-auth only
	FromParams       string `json:"from_params"  xorm:"not null default 'access_token' comment('从指定的URL参数中抽取JWT') VARCHAR(1024)"` // for jwt-auth only   default: ["access_token"]
	FromCookies      string `json:"from_cookies" xorm:"not null default '-' comment('从指定的URL参数中抽取JWT') VARCHAR(4096)"`            // for jwt-auth only   default: -
	KeepToken        string `json:"keep_token" xorm:"not null default 'Y' comment('转发给后端时是否保留JWT') VARCHAR(1)"`
	ClockSkewSeconds string `json:"clock_skew_seconds" xorm:"not null default '60' comment('校验JWT的exp和iat字段时允许的时钟偏移量，单位为秒') VARCHAR(32)"` // for jwt-auth only   default: 60

	// oauth2
	RedirectUrl string `json:"redirect_url" xorm:"oauth2_redirect_url not null default '' comment('oauth2 v1 转发 url') VARCHAR(1024)"`
	// v2
	RedirectUrls string `json:"redirect_urls" xorm:"oauth2_v2_redirect_urls not null default '' comment('oauth2 v2 转发 urls') VARCHAR(4096)"`
	Name         string `json:"name"  xorm:"oauth2_name not null default '' comment('OAuth2 名称') VARCHAR(128)"`
	ClientId     string `json:"client_id" xorm:"oauth2_client_id not null default '' comment('客户端ID') VARCHAR(128)"`
	ClientSecret string `json:"client_secret" xorm:"oauth2_client_secret not null default '' comment('客户端Secret') VARCHAR(128)"`

	BaseRow `xorm:"extends"`
}
