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

package strutil

import (
	"reflect"
)

// ReverseSlice reverses the slice s in place with any type
func ReverseSlice(s interface{}) {
	if reflect.TypeOf(s).Kind() != reflect.Slice {
		return
	}
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func ReverseString(s string) string {
	var n = len(s)
	if n <= 1 {
		return s
	}
	var b = []rune(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

func DedupAnySlice(s interface{}, uniq func(i int) interface{}) interface{} {
	in := reflect.ValueOf(s)
	if in.Kind() != reflect.Slice && in.Kind() != reflect.Array {
		return s
	}
	var (
		dup = make(map[interface{}]struct{})
		out = reflect.MakeSlice(in.Type(), 0, in.Len())
	)
	for i := 0; i < in.Len(); i++ {
		v := uniq(i)
		if _, ok := dup[v]; !ok {
			out = reflect.Append(out, in.Index(i))
			dup[v] = struct{}{}
		}
	}
	return out.Interface()
}

// DistinctArray is used to get unique elements from a slice with comparable type \
func DistinctArray[T comparable](arr []T) []T {
	dupMap := make(map[T]struct{})
	unique := make([]T, 0)
	for _, item := range arr {
		// if not exist, add in the unique array
		if _, ok := dupMap[item]; !ok {
			dupMap[item] = struct{}{}
			unique = append(unique, item)
		}
	}

	return unique
}

// DistinctArrayInStructByFiled it is used to get unique elements from struct array filter by filed
// you should offer a function to get the field from the struct
// besides, the function you can set a skip condition, if skip is true, it will ignore the elem
// it returns the array of Type T(the same as the array offered)
func DistinctArrayInStructByFiled[T any, C comparable](arr []T, fn func(T) (target C, skip bool)) []T {
	dupMap := make(map[C]struct{})
	unique := make([]T, 0)
	for _, item := range arr {
		value, skip := fn(item)
		if skip {
			continue
		}
		if _, ok := dupMap[value]; !ok {
			dupMap[value] = struct{}{}
			unique = append(unique, item)
		}
	}
	return unique
}

// DistinctArrayFiledInStruct it is used to get unique elements from struct
// you should offer a function to get the field from the struct
// besides, the function you can set a skip condition, if skip is true, it will ignore the elem
// it returns the array of Type C(the field type)
func DistinctArrayFiledInStruct[T any, C comparable](arr []T, fn func(T) (target C, skip bool)) []C {
	dupMap := make(map[C]struct{})
	unique := make([]C, 0)
	for _, item := range arr {
		value, skip := fn(item)
		if skip {
			continue
		}
		if _, ok := dupMap[value]; !ok {
			dupMap[value] = struct{}{}
			unique = append(unique, value)
		}
	}
	return unique
}
