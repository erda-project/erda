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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tealeg/xlsx"
)

type TaskSheet struct {
	TaskOnly TaskOnly
}
type TaskOnly struct {
	TaskType     string
	CustomFields []string
}

func Test_Decode(t *testing.T) {
	f, err := os.CreateTemp(".", "test-*.xlsx")
	//defer os.Remove(f.Name())
	assert.NoError(t, err)
	// encode
	cells := [][]Cell{
		// title
		{
			{Value: "TaskOnly", VerticalMergeNum: 0},
			{Value: "TaskOnly", VerticalMergeNum: 0},
		},
		{
			{Value: "TaskType", VerticalMergeNum: 1},
			{Value: "CustomFields", VerticalMergeNum: 0},
		},
		{
			{Value: "TaskType", VerticalMergeNum: 0},
			{Value: "cf-1", VerticalMergeNum: 0},
		},
		// value
		{
			{Value: "code", VerticalMergeNum: 0},
			{Value: "v-of-cf-1", VerticalMergeNum: 0},
		},
	}
	err = ExportExcelByCell(f, cells, "taskonly")
	assert.NoError(t, err)
	// decode
	strCells, err := xlsx.FileToSlice(f.Name())
	assert.NoError(t, err)
	fmt.Printf("%#v\n", strCells)

	// convert [][][]string to map[string][]Cell
	m := make(map[string][]Cell)
	for i := range []int{0, 1} { // column index
		columnName := strings.Join([]string{strCells[0][0][i], strCells[0][1][i], strCells[0][2][i]}, "---")
		// data rows start from 2
		for j := range strCells[0][2:] {
			m[columnName] = append(m[columnName], Cell{
				Value:              strCells[0][j+2][i],
				VerticalMergeNum:   0,
				HorizontalMergeNum: 0,
			})
		}
	}
	fmt.Printf("%#v\n", m)
}

func TestDecodeSheetToSlice(t *testing.T) {
	// test if an Excel file have two sheets with same name
	f := NewFile()
	assert.NoError(t, AddSheetByCell(f, nil, "sheet1"))
	assert.Error(t, AddSheetByCell(f, nil, "sheet1"))

	// decode file with two same-name sheets directly
	// According to the Excel standard, it is not possible to create an Excel file with two sheets of the same name.
}
