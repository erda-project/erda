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

package structparser

import (
	"reflect"
)

// getComment 获取 `t` 的 `fieldname` 对应的注释(comment)
// `t` 是 struct value
func getComment(t interface{}, fieldname string) string {
	tp := reflect.TypeOf(t)
	value := reflect.ValueOf(t)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	method := value.MethodByName("Desc_" + tp.Name())

	if (method == reflect.Value{}) {
		return ""
	}
	return method.Call([]reflect.Value{reflect.ValueOf(fieldname)})[0].String()
}
