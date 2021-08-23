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
