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

package ddlreverser_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/database/sqlparser/ddlreverser"
)

func TestReverseSlice(t *testing.T) {
	var cases = [][]int{
		{1},
		{2, 3},
		{4, 9, 2},
	}
	var results = [][]int{
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
		ddlreverser.ReverseSlice(case_)
		if !equal(case_, results[i]) {
			t.Fatal(i, case_, "fails")
		}
	}
}
