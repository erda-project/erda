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

package arrays_test

import (
	"fmt"

	"github.com/erda-project/erda/pkg/arrays"
)

func ExampleStructArrayToMap() {
	type itemStruct struct {
		Key   string
		Value string
		Other interface{}
	}
	arr := []itemStruct{
		{Key: "this is key1", Value: "this is value1", Other: ""},
		{Key: "this is key2", Value: "this is value2", Other: ""},
		{Key: "", Value: "", Other: ""},
	}

	array2Map := arrays.StructArrayToMap(arr, func(item itemStruct) (key string, value string, skip bool) {
		if item.Key == "" {
			return "", "", true
		}
		return item.Key, item.Value, false
	})

	fmt.Println(array2Map)

	// Output: map[this is key1:this is value1 this is key2:this is value2]
}

func ExampleArrayToMap() {
	arr := []int{1, 2, 3, 4, 3}

	result := arrays.ArrayToMap(arr)
	fmt.Println(result)

	// Output: map[1:{} 2:{} 3:{} 4:{}]
}

func ExampleGetFieldArrFromStruct() {
	type itemStruct struct {
		Field string
	}
	arr := []itemStruct{
		{Field: "123"},
		{Field: "5321"},
		{Field: "45"},
	}

	result := arrays.GetFieldArrFromStruct(arr, func(item itemStruct) string {
		return item.Field
	})

	fmt.Println(result)

	// Output: [123 5321 45]
}
