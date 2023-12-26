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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsContain(t *testing.T) {
	is := IsContain([]string{"a", "b", "c"}, "a")
	assert.Equal(t, is, true)

}

func TestIsArrayContained(t *testing.T) {
	var index int
	var flag bool
	index, flag = IsArrayContained([]int{1, 3, 4, 5}, []int{1, 3, 4})
	assert.Equal(t, index, -1)
	assert.Equal(t, flag, true)

	index, flag = IsArrayContained([]string{"e", "r", "d", "a"}, []string{"e"})
	assert.Equal(t, index, -1)
	assert.Equal(t, flag, true)

	index, flag = IsArrayContained([]uint64{1}, []uint64{})
	assert.Equal(t, index, -1)
	assert.Equal(t, flag, true)

	index, flag = IsArrayContained([]int64{1, 3, 5, 2}, []int64{2, 5, 4})
	assert.Equal(t, index, 2)
	assert.Equal(t, flag, false)

	index, flag = IsArrayContained([]uint8{}, []uint8{'a'})
	assert.Equal(t, index, 0)
	assert.Equal(t, flag, false)
}

func TestDifferenceSet(t *testing.T) {
	arr1 := []uint64{1, 2, 3, 4, 5}
	arr2 := []uint64{2, 4}

	result := DifferenceSet(arr1, arr2)

	expectedResult := []uint64{1, 3, 5}
	assert.Equal(t, expectedResult, result)
}

func TestArray2Map(t *testing.T) {
	arr := []int{1, 2, 3, 4, 3}
	want := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
	}

	result := Array2Map(arr)
	assert.Equal(t, want, result)
}

func TestStructArray2Map(t *testing.T) {
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

	want := map[string]string{
		"this is key1": "this is value1",
		"this is key2": "this is value2",
	}

	array2Map := StructArray2Map(arr, func(item itemStruct) (key string, value string, skip bool) {
		if item.Key == "" {
			return "", "", true
		}
		return item.Key, item.Value, false
	})

	assert.Equal(t, want, array2Map)

}

func TestGetFieldArrFromStruct(t *testing.T) {
	type itemStruct struct {
		Field string
	}
	arr := []itemStruct{
		{Field: "123"},
		{Field: "5321"},
		{Field: "45"},
	}
	want := []string{"123", "5321", "45"}
	result := GetFieldArrFromStruct(arr, func(item itemStruct) string {
		return item.Field
	})
	assert.Equal(t, want, result)
}
