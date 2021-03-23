package template

import "regexp"

// Render 渲染简单模板 占位符格式: {{key}}
func Render(template string, params map[string]string) string {
	reg := regexp.MustCompile(`{{.+?}}`)
	result := reg.ReplaceAllStringFunc(template, func(s string) string {
		key := s[2 : len(s)-2]
		value, ok := params[key]
		if ok {
			return value
		}
		return s
	})
	return result
}
