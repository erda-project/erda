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
