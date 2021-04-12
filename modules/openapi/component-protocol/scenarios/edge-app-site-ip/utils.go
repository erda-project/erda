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

package edgeappsiteip

import (
	"fmt"
	"reflect"
	"strconv"
)

func StructToMap(obj interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		fmt.Println("it is not struct")
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		if tagVal := t.Field(i).Tag.Get("json"); tagVal != "" {
			out[tagVal] = v.Field(i).Interface()
		}
	}
	return out
}

func Force2Int(input interface{}) (int, error) {
	var res int
	if m, ok := input.(string); ok {
		res, err := strconv.Atoi(m)
		if err != nil {
			return res, err
		}
		return res, nil
	}
	if m, ok := input.(int64); ok {
		res = int(m)
		return res, nil
	}
	if m, ok := input.(float64); ok {
		res = int(m)
		return res, nil
	}

	return res, nil
}
