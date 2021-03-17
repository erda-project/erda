package i18n

import (
	"fmt"
	"regexp"
)

type Template struct {
	key     string
	content string
}

func NewTemplate(key string, content string) *Template {
	return &Template{key: key, content: content}
}

func (t *Template) Render(args ...interface{}) string {
	return fmt.Sprintf(t.content, args...)
}

func (t *Template) Key() string {
	return t.key
}

func (t *Template) Content() string {
	return t.content
}

// RenderByKey 根据key名字渲染  eg: {{keyName}}
func (template *Template) RenderByKey(params map[string]string) string {
	reg := regexp.MustCompile(`{{.+?}}`)
	result := reg.ReplaceAllStringFunc(template.content, func(s string) string {
		key := s[2 : len(s)-2]
		value, ok := params[key]
		if ok {
			return value
		}
		return s
	})
	return result
}
