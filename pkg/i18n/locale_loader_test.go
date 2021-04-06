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
	"testing"
)

func TestLocaleLoader(t *testing.T) {
	loader := NewLoader()
	err := loader.LoadFile("test-locale1.json", "test-locale2.json", "test-locale.yml")
	if err != nil {
		panic(err)
	}
	println(loader.Locale(ZH).Get("dice.not_login"))
	println(loader.Locale(EN).Get("dice.not_login"))
	println(loader.Locale(ZH).Get("dice.multiline"))

}

func TestLocaleLoaderTemplate(t *testing.T) {
	loader := NewLoader()
	err := loader.LoadFile("test-locale1.json", "test-locale2.json", "test-locale.yml")
	if err != nil {
		panic(err)
	}
	template := loader.Locale(ZH).GetTemplate("dice.resource_not_found")
	println(template.RenderByKey(map[string]string{
		"name": "11",
	}))
}
