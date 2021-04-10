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

package custom

import (
	"fmt"
	"regexp"

	"github.com/erda-project/erda/modules/hepa/apipolicy"
)

type PolicyDto struct {
	apipolicy.BaseDto
	Switch         bool     `json:"switch"`
	UserCookie     string   `json:"userCookie"`
	ExcludedMethod []string `json:"excludedMethod"`
	TokenName      string   `json:"tokenName"`
	TokenDomain    string   `json:"tokenDomain"`
	CookieSecure   bool     `json:"cookieSecure"`
	ValidTTL       int64    `json:"validTTL"`
	RefreshTTL     int64    `json:"refreshTTL"`
	ErrStatus      int64    `json:"errStatus"`
	ErrMsg         string   `json:"errMsg"`
}

func (dto PolicyDto) IsValidDto() (bool, string) {
	if !dto.Switch {
		return true, ""
	}
	if dto.UserCookie == "" {
		return false, "鉴别用户的cookie名称不能为空"
	}
	if dto.TokenName == "" {
		return false, "token的名称不能为空"
	}
	if dto.TokenDomain != "" {
		if ok, _ := regexp.MatchString(`^[0-9a-zA-z-_\.:]+$`, dto.TokenDomain); !ok {
			return false, fmt.Sprintf("cookie生效域名字段不合法:%s", dto.TokenDomain)
		}
	}
	if dto.ValidTTL <= 0 {
		return false, "token过期时间需要大于0"
	}
	if dto.RefreshTTL <= 0 {
		return false, "token更新周期需要大于0"
	}
	if dto.ErrStatus < 100 || dto.ErrStatus >= 600 {
		return false, "请填写合法的校验失败状态码"
	}
	if dto.ErrMsg == "" {
		return false, "校验失败应答不能为空"
	}
	return true, ""
}
