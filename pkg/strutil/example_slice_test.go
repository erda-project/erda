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
	"crypto/sha256"
	"fmt"

	"github.com/erda-project/erda/pkg/strutil"
)

func ExampleDistinctArrayInStructByFiled() {
	type TestIntStruct struct {
		Name  string
		Field int
	}

	input := []TestIntStruct{
		{Field: 1, Name: "1"},
		{Field: 2, Name: "2"},
		{Field: 3, Name: "3"},
		{Field: 3, Name: "3"},
	}

	output := strutil.DistinctArrayInStructByFiled(input, func(t TestIntStruct) (string, bool) {
		hash := sha256.Sum256([]byte(fmt.Sprintf("%s%d", t.Name, t.Field)))
		return string(hash[:]), false
	})

	fmt.Println(output)

	// Output: [{1 1} {2 2} {3 3}]
}

func ExampleDistinctArrayFiledInStruct() {
	type TestIntStruct struct {
		Name  string
		Field int
	}

	input := []TestIntStruct{
		{Field: 1, Name: "1"},
		{Field: 2, Name: "2"},
		{Field: 1, Name: "3"},
	}

	output := strutil.DistinctArrayFiledInStruct(input, func(t TestIntStruct) (int, bool) {
		return t.Field, false
	})

	fmt.Println(output)

	// Output: [1 2]
}
