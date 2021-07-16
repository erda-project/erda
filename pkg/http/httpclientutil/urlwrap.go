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

package httpclientutil

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	httpProto  = "http://"
	httpsProto = "https://"
)

func WrapHttp(addr string) string {
	if strings.HasPrefix(addr, httpProto) {
		return addr
	}
	if strings.HasPrefix(addr, httpsProto) {
		return httpProto + strings.TrimLeft(addr, httpsProto)
	}
	return httpProto + addr
}

func WrapHttps(addr string) string {
	if strings.HasPrefix(addr, httpsProto) {
		return addr
	}
	if strings.HasPrefix(addr, httpProto) {
		return httpsProto + strings.TrimLeft(addr, httpProto)
	}
	return httpsProto + addr
}

func WrapProto(addr string) string {
	if strings.HasPrefix(addr, httpsProto) || strings.HasPrefix(addr, httpProto) {
		return addr
	}

	// default return http address
	return httpProto + addr
}

func RmProto(url string) string {
	if strings.HasPrefix(url, httpProto) {
		return strings.TrimPrefix(url, httpProto)
	}
	if strings.HasPrefix(url, httpsProto) {
		return strings.TrimPrefix(url, httpsProto)
	}
	return url
}

func GetHost(rawurl string) string {
	URL, _ := url.Parse(WrapProto(rawurl))
	return URL.Host
}

func GetPath(rawurl string) string {
	URL, _ := url.Parse(WrapProto(rawurl))
	return URL.Path
}

func GetPort(rawurl string) int {
	URL, _ := url.Parse(WrapProto(rawurl))
	s := URL.Port()
	if s == "" {
		return 80
	}
	i, _ := strconv.Atoi(s)
	return i
}

func CombineHostAndPath(host string, path string) string {
	host = WrapProto(host)
	if strings.HasSuffix(host, "/") {
		host = host[:len(host)-1]
	}
	return host + path
}

func HasSchema(rawurl string) bool {
	if strings.HasPrefix(rawurl, httpProto) || strings.HasPrefix(rawurl, httpsProto) {
		return true
	}
	return false
}
