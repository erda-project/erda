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

type LocaleResource struct {
	name               string
	resourceMap        map[string]string
	defaultResourceMap map[string]string
}

func (locale *LocaleResource) Name() string {
	return locale.name
}

func (locale *LocaleResource) ExistKey(key string) bool {
	_, ok := locale.resourceMap[key]
	if ok {
		return true
	}
	if locale.defaultResourceMap != nil {
		_, ok := locale.defaultResourceMap[key]
		return ok
	}
	return false
}

func (locale *LocaleResource) Get(key string, defaults ...string) string {
	content, ok := locale.resourceMap[key]
	if !ok {
		// 不存在尝试使用默认
		if locale.defaultResourceMap != nil {
			content, ok := locale.defaultResourceMap[key]
			if !ok {
				if len(defaults) > 0 {
					return defaults[0]
				}
				return key
			}
			return content
		}
		if len(defaults) > 0 {
			return defaults[0]
		}
		return key
	}
	return content
}

func (locale *LocaleResource) GetTemplate(key string) *Template {
	content := locale.Get(key)
	return NewTemplate(key, content)
}
