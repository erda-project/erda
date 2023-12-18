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

package excel

import (
	"os"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func Test_ensureDropList1(t *testing.T) {
	assert.Equal(t, 3, utf8.RuneCountInString("\"a\""))

	ss := []string{"a", "b", "c"}
	ss = ss[:len(ss)]

	dropList := []string{"a", "b", "c"} // 3(a)+2(,)+2(")=7
	ensureDropList(&dropList, 8)
	assert.Equal(t, 3, len(dropList))

	dropList = []string{"a", "b", "c"}
	ensureDropList(&dropList, 7)
	assert.Equal(t, 3, len(dropList))

	dropList = []string{"a", "b", "c"}
	ensureDropList(&dropList, 6)
	assert.Equal(t, 2, len(dropList))

	dropList = []string{"a", "b", "c"} // 2(a)+1(,)+2(")=5
	ensureDropList(&dropList, 5)
	assert.Equal(t, 2, len(dropList))

	dropList = []string{"a", "b", "c"} // 2+1+2=5
	ensureDropList(&dropList, 4)
	assert.Equal(t, 1, len(dropList))

	dropList = []string{"a", "b", "c"}
	ensureDropList(&dropList, 3)
	assert.Equal(t, 1, len(dropList))

	dropList = []string{"a", "b", "c"} // 1(a)+2(")=3
	ensureDropList(&dropList, 2)
	assert.Equal(t, 0, len(dropList))

	dropList = []string{"a", "b", "c"} // 1(a)+2(")=3
	ensureDropList(&dropList, 1)
	assert.Equal(t, 0, len(dropList))

	dropList = []string{"a", "b", "c"} // 1(a)+2(")=3
	ensureDropList(&dropList, 0)
	assert.Equal(t, 0, len(dropList))
}

func TestExcelDataValidationError(t *testing.T) {
	f := NewFile()
	row1 := []Cell{NewTitleCell("status")}
	row2 := []Cell{NewCell("")}
	data := [][]Cell{row1, row2}
	err := AddSheetByCell(f, data, "test", NewSheetHandlerForDropList(1, 0, 1, 0, []string{"", "init", "ing", "done"}))
	outputF, err := os.Create("./test_with_error_style.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outputF.Name())
	err = WriteFile(outputF, f, "test_with_error_style.xlsx")
	if err != nil {
		t.Fatal(err)
	}
}
