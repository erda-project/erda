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

package cors

import (
	"github.com/erda-project/erda/modules/hepa/apipolicy"
)

const (
	ANNOTATION_CORS_ENABLE      = "nginx.ingress.kubernetes.io/enable-cors"
	ANNOTATION_CORS_METHODS     = "nginx.ingress.kubernetes.io/cors-allow-methods"
	ANNOTATION_CORS_HEADERS     = "nginx.ingress.kubernetes.io/cors-allow-headers"
	ANNOTATION_CORS_ORIGIN      = "nginx.ingress.kubernetes.io/cors-allow-origin"
	ANNOTATION_CORS_CREDENTIALS = "nginx.ingress.kubernetes.io/cors-allow-credentials"
	ANNOTATION_CORS_MAXAGE      = "nginx.ingress.kubernetes.io/cors-max-age"
)

type PolicyDto struct {
	apipolicy.BaseDto
	Methods     string `json:"methods"`
	Headers     string `json:"headers"`
	Origin      string `json:"origin"`
	Credentials bool   `json:"credentials"`
	MaxAge      int64  `json:"maxAge"`
}

func (dto PolicyDto) IsValidDto() (bool, string) {
	if !dto.Switch {
		return true, ""
	}
	if dto.Methods == "" {
		return false, "HTTP方法字段不能为空"
	}
	if dto.Methods == "*" {
		return false, "HTTP方法不允许配置通配符"
	}
	if dto.Headers == "" {
		return false, "HTTP请求头字段不能为空"
	}
	if dto.Headers == "*" {
		return false, "HTTP请求头不允许配置通配符"
	}
	if dto.Origin == "" {
		return false, "跨域地址字段不能为空"
	}
	if dto.MaxAge <= 0 {
		return false, "预检请求缓存时间需要大于0"
	}
	return true, ""
}
