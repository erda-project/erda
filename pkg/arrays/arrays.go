// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arrays

func Distinct(array []string) []string {
	data := make(map[string]interface{})
	for _, v := range array {
		data[v] = nil
	}
	var result []string
	for k := range data {
		result = append(result, k)
	}
	return result
}

func Concat(array []string, arrays ...[]string) []string {
	for _, arr := range arrays {
		array = append(array, arr...)
	}
	return array
}

func IsContain(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func Paging(pageNo, pageSize, length uint64) (int64, int64) {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > length {
		pageSize = length
	}
	from, end := (pageNo-1)*pageSize, pageNo*pageSize
	if from > length {
		return -1, -1
	}
	if end > length {
		end = length
	}
	return int64(from), int64(end)
}

// IsArrayContained
// Check if the elements in the `sub` array are a subset of the `array`
// It returns (-1,true) if all elements in `sub` are present int `array`
// Otherwise it returns first elements index in `sub` which is not a subset of the `array` and false
func IsArrayContained[T comparable](array []T, sub []T) (int, bool) {
	if len(sub) == 0 {
		return -1, true
	}
	if len(array) == 0 {
		return 0, false
	}
	arrayMap := make(map[T]struct{})
	for _, item := range array {
		arrayMap[item] = struct{}{}
	}

	for index, item := range sub {
		if _, ok := arrayMap[item]; !ok {
			return index, false
		}
	}

	return -1, true
}

func DifferenceSet[T comparable](arr1, arr2 []T) []T {
	arr2Map := make(map[T]bool)
	for _, v := range arr2 {
		arr2Map[v] = true
	}

	// 原地修改 arr1，移除存在于 arr2 中的元素
	j := 0
	for i := 0; i < len(arr1); i++ {
		if !arr2Map[arr1[i]] {
			arr1[j] = arr1[i]
			j++
		}
	}
	arr1 = arr1[:j]
	return arr1
}

// StructArrayToMap converts a struct array into a map after deduplication
// and you should offer a fn to get the key,value and ifSkip the kvs from the struct.
func StructArrayToMap[T any, Key comparable, Value comparable](arr []T, getKV func(T) (Key, Value, bool)) map[Key]Value {
	arr2Map := make(map[Key]Value)

	for _, item := range arr {
		key, value, skip := getKV(item)
		if skip {
			continue
		}
		if _, ok := arr2Map[key]; ok {
			continue
		}
		arr2Map[key] = value
	}

	return arr2Map
}

// ArrayToMap convert the deduplicated array into map
func ArrayToMap[Key comparable](keys []Key) map[Key]struct{} {
	arrToMap := make(map[Key]struct{})

	for _, key := range keys {
		if _, ok := arrToMap[key]; ok {
			continue
		}
		arrToMap[key] = struct{}{}
	}
	return arrToMap
}

// GetFieldArrFromStruct
// construct a elem array from the structs
func GetFieldArrFromStruct[Struct any, Filed any](s []Struct, getField func(Struct) Filed) []Filed {
	arr := make([]Filed, len(s))
	for index, item := range s {
		arr[index] = getField(item)
	}
	return arr
}
