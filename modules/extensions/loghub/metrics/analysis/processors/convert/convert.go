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

package convert

import "strconv"

// Converters .
var Converters = map[string]func(text string) (interface{}, error){
	"string": func(text string) (interface{}, error) {
		return text, nil
	},
	"number": func(text string) (interface{}, error) {
		return strconv.ParseFloat(text, 64)
	},
	"bool": func(text string) (interface{}, error) {
		return strconv.ParseBool(text)
	},
	"timestamp": func(text string) (interface{}, error) {
		return strconv.ParseInt(text, 10, 64)
	},
	"int": func(text string) (interface{}, error) {
		return strconv.ParseInt(text, 10, 64)
	},
	"float": func(text string) (interface{}, error) {
		return strconv.ParseFloat(text, 64)
	},
}

// Converter .
func Converter(typ string) func(text string) (interface{}, error) {
	c, ok := Converters[typ]
	if !ok {
		return Converters["string"]
	}
	return c
}
