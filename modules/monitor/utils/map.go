// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package utils

import "strings"

func GetMapValueString(m map[string]interface{}, key string) (string, bool) {
	value, ok := m[key]
	if !ok {
		return "", false
	}
	return ConvertString(value)
}

func GetMapValueBool(m map[string]interface{}, key string) (bool, bool) {
	value, ok := m[key]
	if !ok {
		return false, false
	}
	return ConvertBool(value)
}

func GetMapValueInt64(m map[string]interface{}, key string) (int64, bool) {
	value, ok := m[key]
	if !ok {
		return 0, false
	}
	return ConvertInt64(value)
}

func GetMapValueUint64(m map[string]interface{}, key string) (uint64, bool) {
	value, ok := m[key]
	if !ok {
		return 0, false
	}
	return ConvertUint64(value)
}

func GetMapValueFloat64(m map[string]interface{}, key string) (float64, bool) {
	value, ok := m[key]
	if !ok {
		return 0, false
	}
	return ConvertFloat64(value)
}

func GetMapValueArr(m map[string]interface{}, key string) ([]interface{}, bool) {
	value, ok := m[key]
	if !ok {
		return nil, false
	}
	arr, ok := value.([]interface{})
	return arr, ok
}

func GetMapValueMap(m map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, ok := m[key]
	if !ok {
		return nil, false
	}
	mm, ok := value.(map[string]interface{})
	return mm, ok
}

func GetMapValue(key string, data map[string]interface{}) interface{} {
	keys := strings.Split(key, ".")
	last := len(keys) - 1
	for i, k := range keys {
		if i >= last {
			return data[k]
		}
		d, ok := data[k]
		if !ok {
			return nil
		}
		data, ok = d.(map[string]interface{})
		if !ok {
			return nil
		}
	}
	return nil
}
