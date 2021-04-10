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

package proxy

import (
	"github.com/erda-project/erda/modules/hepa/apipolicy"
)

const (
	ANNOTATION_REQ_BUFFER         = "nginx.ingress.kubernetes.io/proxy-request-buffering"
	ANNOTATION_RESP_BUFFER        = "nginx.ingress.kubernetes.io/proxy-buffering"
	ANNOTATION_REQ_LIMIT          = "nginx.ingress.kubernetes.io/proxy-body-size"
	ANNOTATION_PROXY_REQ_TIMEOUT  = "nginx.ingress.kubernetes.io/proxy-send-timeout"
	ANNOTATION_PROXY_RESP_TIMEOUT = "nginx.ingress.kubernetes.io/proxy-read-timeout"
	ANNOTATION_SSL_REDIRECT       = "nginx.ingress.kubernetes.io/ssl-redirect"
)

type PolicyDto struct {
	apipolicy.BaseDto
	ReqBuffer         bool  `json:"reqBuffer"`
	RespBuffer        bool  `json:"respBuffer"`
	ClientReqLimit    int64 `json:"clientReqLimit"`
	ClientReqTimeout  int64 `json:"clientReqTimeout"`
	ClientRespTimeout int64 `json:"clientRespTimeout"`
	ProxyReqTimeout   int64 `json:"proxyReqTimeout"`
	ProxyRespTimeout  int64 `json:"proxyRespTimeout"`
	HostPassthrough   bool  `json:"hostPassthrough"`
	SSLRedirect       bool  `json:"sslRedirect"`
}

func (dto PolicyDto) IsValidDto() (bool, string) {
	if !dto.Switch {
		return true, ""
	}
	if dto.ClientReqLimit <= 0 {
		return false, "客户端请求限制需要大于0"
	}
	if dto.ClientReqTimeout <= 0 {
		return false, "客户端请求超时需要大于0"
	}
	if dto.ClientRespTimeout <= 0 {
		return false, "客户端应答超时需要大于0"
	}
	if dto.ProxyReqTimeout <= 0 {
		return false, "后端请求超时需要大于0"
	}
	if dto.ProxyRespTimeout <= 0 {
		return false, "后端应答超时需要大于0"
	}
	return true, ""
}
