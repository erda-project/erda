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
