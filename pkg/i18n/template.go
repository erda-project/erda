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
