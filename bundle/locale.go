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
