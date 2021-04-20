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

package i18n

import "net/http"

const ZH = "zh-CN"
const EN = "en-US"

// GetLocaleNameByRequest 从request获取语言名称
func GetLocaleNameByRequest(request *http.Request) string {
	// 优先querystring 其次header
	lang := request.URL.Query().Get("lang")
	if lang != "" {
		return lang
	}
	lang = request.Header.Get("Lang")
	if lang != "" {
		return lang
	}
	return ""
}
