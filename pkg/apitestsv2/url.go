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

package apitestsv2

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/erda-project/erda/pkg/httpclientutil"
)

// domain: schema://${prefix} + u
func polishURL(rawurl string, domain string) (string, error) {
	if rawurl == "" {
		return "", fmt.Errorf("empty url")
	}

	// 将 URL 后面的参数去掉
	rawurl = strings.Split(rawurl, "?")[0]

	// 若 url 非 http://, https:// 开头，则补充 domain
	if !httpclientutil.HasSchema(rawurl) {
		rawurl = domain + rawurl
	}

	// 保证协议头存在
	rawurl = httpclientutil.WrapProto(rawurl)

	if _, err := url.Parse(rawurl); err != nil {
		return "", fmt.Errorf("invalid url: %s, err: %v", rawurl, err)
	}

	return rawurl, nil
}
