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

package vars

import (
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	I18nLang_zh_CN = "zh-CN"
	I18nLang_en_US = "en-US"
)

func (data *DataForFulfill) I18n(key string, args ...interface{}) string {
	return doI18n(data.Tran, data.Lang, key, args...)
}

func doI18n(tran i18n.Translator, lang i18n.LanguageCodes, key string, args ...interface{}) string {
	if len(args) == 0 {
		try := tran.Text(lang, key)
		if try != key {
			return try
		}
	}
	return tran.Sprintf(lang, key, args...)
}

func (data *DataForFulfill) AllI18nValuesByKey(key string, args ...interface{}) []string {
	// iterate all supported langs
	langs := allSupportedLangs()
	var allValues []string
	for _, lang := range langs {
		s := doI18n(data.Tran, lang, key, args...)
		allValues = append(allValues, s)
	}
	return strutil.DedupSlice(allValues, true)
}

func GetI18nLang(locale string) i18n.LanguageCodes {
	lang, _ := i18n.ParseLanguageCode(locale)
	if lang.Len() > 0 {
		return lang
	}
	l, _ := i18n.ParseLanguageCode(I18nLang_zh_CN)
	return l
}

func allSupportedLangs() []i18n.LanguageCodes {
	l1, _ := i18n.ParseLanguageCode(I18nLang_zh_CN)
	l2, _ := i18n.ParseLanguageCode(I18nLang_en_US)
	return []i18n.LanguageCodes{l1, l2}
}
