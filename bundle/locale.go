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

package bundle

import (
	"net/http"

	"github.com/erda-project/erda/pkg/i18n"
)

// GetLocale 获取对应语言对象
func (b *Bundle) GetLocale(locales ...string) *i18n.LocaleResource {
	return b.i18nLoader.Locale(locales...)
}

// GetLocaleByRequest 从request获取语言对象
func (b *Bundle) GetLocaleByRequest(request *http.Request) *i18n.LocaleResource {
	locale := i18n.GetLocaleNameByRequest(request)
	return b.i18nLoader.Locale(locale)
}

func (b *Bundle) GetLocaleLoader() *i18n.LocaleResourceLoader {
	return b.i18nLoader
}
