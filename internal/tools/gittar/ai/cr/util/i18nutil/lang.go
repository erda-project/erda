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

package i18nutil

import (
	"strings"

	"github.com/erda-project/erda-infra/providers/i18n"
)

const (
	Chinese = "Chinese"
	English = "English"

	CodeZh = "zh"
	CodeEn = "en"
)

func GetUserLang(langs i18n.LanguageCodes) string {
	var code string
	if len(langs) == 0 {
		code = CodeZh
	} else {
		code = langs[0].RestrictedCode()
	}
	code = strings.ToLower(code)
	switch code {
	case CodeZh:
		return Chinese
	case CodeEn:
		return English
	default:
		return Chinese
	}
}
