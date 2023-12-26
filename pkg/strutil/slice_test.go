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

package strutil_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/strutil"
)

func TestReverseSlice(t *testing.T) {
	var cases = [][]int{
		nil,
		{},
		{1},
		{2, 3},
		{4, 9, 2},
	}
	var results = [][]int{
		nil,
		{},
		{1},
		{3, 2},
		{2, 9, 4},
	}

	equal := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	for i, case_ := range cases {
		strutil.ReverseSlice(case_)
		if !equal(case_, results[i]) {
			t.Fatal(i, case_, "fails")
		}
	}

	strutil.ReverseSlice([2]int{0, 2})
	strutil.ReverseSlice("a")
}

func TestDedupAnySlice(t *testing.T) {
	var (
		a = []int{2, 3, 4, 6, 7, 8, 9, 7, 3, 6}
		b = []string{"json", "yaml", "txt", "yaml"}
		c = []struct {
			Name string
			Age  int
		}{{Name: "dspo", Age: 13}, {Name: "cmc", Age: 14}, {Name: "yuxiaoer", Age: 15}, {Name: "cmc", Age: 14}}
	)
	a = strutil.DedupAnySlice(a, func(i int) interface{} {
		return a[i]
	}).([]int)
	b = strutil.DedupAnySlice(b, func(i int) interface{} {
		return b[i]
	}).([]string)
	c = strutil.DedupAnySlice(c, func(i int) interface{} {
		return c[i].Name
	}).([]struct {
		Name string
		Age  int
	})
	t.Log(a)
	t.Log(b)
	t.Log(c)
}

func TestReverseString(t *testing.T) {
	var (
		s1 = "desrever si siht"
		s2 = "a"
	)
	if reversed := strutil.ReverseString(s1); reversed != "this is reversed" {
		t.Error("error")
	}
	if reversed := strutil.ReverseString(s2); reversed != s2 {
		t.Error("error")
	}
}

func TestDistinctArray(t *testing.T) {
	var (
		a      = []string{"1", "@", "#", "@", "3", "124", "@@", "#"}
		a_want = []string{"1", "@", "#", "3", "124", "@@"}
		b      = []int{1, 5, 2, 5, 5, 7}
		b_want = []int{1, 5, 2, 7}
	)

	a = strutil.DistinctArray(a)
	for i := 0; i < len(a); i++ {
		if a[i] != a_want[i] {
			t.Error("err in distinctArrayWithFilter in array 'a'")
		}
	}

	b = strutil.DistinctArray(b)
	for i := 0; i < len(b); i++ {
		if b[i] != b_want[i] {
			t.Error("err in distinctArrayWithFilter in array 'b'")
		}
	}

}

func TestDistinctArrayInStructByFiled(t *testing.T) {
	type TestIntStruct struct {
		Name  string
		Field int
	}

	type TestStringStruct struct {
		Name  string
		Field string
	}

	testInt := []TestIntStruct{
		{Field: 1, Name: "1"},
		{Field: 2, Name: "2"},
		{Field: 1, Name: "3"},
	}
	testIntWant := []TestIntStruct{
		{Field: 1, Name: "1"},
		{Field: 2, Name: "2"},
	}

	testString := []TestStringStruct{
		{Name: "1", Field: "123"},
		{Name: "2", Field: "234"},
		{Name: "#", Field: "#"},
		{Name: "#2", Field: "#"},
		{Name: "#3", Field: "#"},
		{Name: "123", Field: "123"},
	}

	testStringWant := []TestStringStruct{
		{Name: "2", Field: "234"},
		{Name: "#", Field: "#"},
	}

	testIntResp := strutil.DistinctArrayInStructByFiled(testInt, func(t TestIntStruct) (int, bool) {
		return t.Field, false
	})
	if len(testIntResp) != len(testIntWant) {
		t.Error("length is not equal")
	}
	for i := 0; i < len(testIntWant); i++ {
		if testIntWant[i] != testIntResp[i] {
			t.Error("err in test int")
			return
		}
	}

	testStringResp := strutil.DistinctArrayInStructByFiled(testString, func(t TestStringStruct) (string, bool) {
		if t.Field == "123" {
			return t.Field, true
		}
		return t.Field, false
	})
	if len(testStringResp) != len(testStringWant) {
		t.Error("length is not equal")
	}
	for i := 0; i < len(testStringResp); i++ {
		if testStringWant[i] != testStringResp[i] {
			t.Error("err in test string")
			return
		}
	}

}

func TestDistinctArrayFiledInStruct(t *testing.T) {
	type TestIntStruct struct {
		Name  string
		Field int
	}

	type TestStringStruct struct {
		Name  string
		Field string
	}

	testInt := []TestIntStruct{
		{Field: 1, Name: "1"},
		{Field: 2, Name: "2"},
		{Field: 1, Name: "3"},
	}
	testIntWant := []int{1, 2}

	testString := []TestStringStruct{
		{Name: "1", Field: "123"},
		{Name: "2", Field: "234"},
		{Name: "#", Field: "#"},
		{Name: "#2", Field: "#"},
		{Name: "#3", Field: "#"},
		{Name: "123", Field: "123"},
	}

	testStringWant := []string{"123", "234"}

	testIntResp := strutil.DistinctArrayFiledInStruct(testInt, func(t TestIntStruct) (int, bool) {
		return t.Field, false
	})
	if len(testIntResp) != len(testIntWant) {
		t.Error("length is not equal")
	}
	for i := 0; i < len(testIntWant); i++ {
		if testIntWant[i] != testIntResp[i] {
			t.Error("err in test int")
			return
		}
	}

	testStringResp := strutil.DistinctArrayFiledInStruct(testString, func(t TestStringStruct) (string, bool) {
		if t.Field == "#" {
			return t.Field, true
		}
		return t.Field, false
	})
	if len(testStringResp) != len(testStringWant) {
		t.Error("length is not equal")
	}
	for i := 0; i < len(testStringResp); i++ {
		if testStringWant[i] != testStringResp[i] {
			t.Error("err in test string")
			return
		}
	}

}
