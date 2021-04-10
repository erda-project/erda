// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package orm

import (
	kongDto "github.com/erda-project/erda/modules/hepa/kong/dto"
)

const (
	OAUTH2       = "oauth2"
	KEYAUTH      = "key-auth"
	SIGNAUTH     = "sign-auth"
	HMACAUTH     = "hmac-auth"
	KeyAuthTips  = "请将appKey带在名为appKey的url参数或者名为X-App-Key的请求头上"
	SignAuthTips = "请将appKey带在名为appKey的url参数上，将参数签名串带在名为sign的url参数上"
)

type AuthItem struct {
	AuthType string                        `json:"authType"`
	AuthData kongDto.KongCredentialListDto `json:"authData"`
	AuthTips string                        `json:"authTips"`
}

type ConsumerAuthConfig struct {
	Auths []AuthItem `json:"auths"`
}
