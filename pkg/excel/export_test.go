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

	"github.com/stretchr/testify/assert"
	"github.com/tealeg/xlsx/v3"
)

func TestExportExcelByCell(t *testing.T) {
	file, err := os.OpenFile("/tmp/test1.xlsx", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	assert.NoError(t, err)
	var data [][]Cell
	// row1
	row1 := []Cell{
		NewVMergeCell("用例编号", 1),
		NewVMergeCell("用例名称", 1),
		NewVMergeCell("测试集", 1),
		NewVMergeCell("优先级", 1),
		NewVMergeCell("前置条件", 1),
		NewHMergeCell("步骤与结果", 1),
		EmptyCell(),
		NewHMergeCell("接口测试", 7),
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
	}
	row2 := []Cell{
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
		EmptyCell(),
		NewCell("操作步骤"),
		NewCell("预期结果"),
		NewCell("接口名称"),
		NewCell("请求头信息"),
		NewCell("方法"),
		NewCell("接口地址"),
		NewCell("接口参数"),
		NewCell("请求体"),
		NewCell("out 参数"),
		NewCell("断言"),
	}
	data = append(data, row1, row2)
	err = ExportExcelByCell(file, data, "测试用例")
	assert.NoError(t, err)
}

func TestExportExcelByRowStruct(t *testing.T) {
	file, err := os.OpenFile("/tmp/test2.xlsx", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	assert.NoError(t, err)
	var data [][]Cell
	// row1
	row1 := []Cell{
		NewVMergeCell("RequirementOnly", 1),
	}
	row2 := []Cell{
		NewCell("test"),
	}
	data = append(data, row1, row2)
	err = ExportExcelByCell(file, data, "测试用例")
	assert.NoError(t, err)
}

func TestExportWithDataValidation(t *testing.T) {
	f := NewFile()
	requirementStyle := DefaultTitleCellStyle()
	requirementStyle.Fill.FgColor = "FF16C2C2"
	taskStyle := DefaultTitleCellStyle()
	taskStyle.Fill.FgColor = "FF697FFF"
	bugStyle := DefaultTitleCellStyle()
	bugStyle.Fill.FgColor = "FFF3B519"
	cells := [][]Cell{
		{
			NewTitleCell("ID"),
			NewTitleCell("类型", WithMergeNum(0, 1)),
			NewTitleCell("需求专属字段", WithTitleStyle(requirementStyle)),
			NewTitleCell("任务专属字段", WithTitleStyle(taskStyle)),
			NewTitleCell("缺陷专属字段", WithTitleStyle(bugStyle)),
		},
		{
			NewTitleCell("ID"),
			NewTitleCell("类型"),
			NewTitleCell("需求专属字段", WithTitleStyle(requirementStyle)),
			NewTitleCell("任务专属字段", WithTitleStyle(taskStyle)),
			NewTitleCell("缺陷专属字段", WithTitleStyle(bugStyle)),
		},
		{
			NewCell("10000000001"),
			NewCell("需求"),
			NewCell("hello"),
			NewCell("world"),
			NewCell("owner"),
		},
		{
			NewCell("10000000002"),
			NewCell("任务"),
			NewCell("hello"),
			NewCell("world"),
			NewCell("owner"),
		},
	}
	err := AddSheetByCell(f, cells, "issue",
		NewSheetHandlerForDropList(2, 1, 4, 1, []string{"需求", "任务", "缺陷"}),
		NewSheetHandlerForAutoColWidth(len(cells[1])),
	)
	assert.NoError(t, err)
	outF, err := os.Create("./testdata/cell_dv.xlsx")
	assert.NoError(t, err)
	err = WriteFile(outF, f, "cell_dv")
	assert.NoError(t, err)
}

func TestSheetValidation(t *testing.T) {
	xlsxF := xlsx.NewFile()
	sheet, err := xlsxF.AddSheet("test")
	assert.NoError(t, err)

	row1 := sheet.AddRow()
	cell1_1 := row1.AddCell()
	cell1_1.Value = "type"
	row2 := sheet.AddRow()
	cell2_1 := row2.AddCell()
	cell2_1.Value = "type"

	// merge row1 and row2
	cell1_1.VMerge = 1

	// add droplist for cell3_1
	row3 := sheet.AddRow()
	cell3_1 := row3.AddCell()
	cell3_1.Value = "bug"

	dv := xlsx.NewDataValidation(2, 0, 2, 0, true)
	err = dv.SetDropList([]string{"bug", "task"})
	assert.NoError(t, err)
	title := "提示"
	msg := "试试"
	dv.SetInput(&title, &msg)
	sheet.AddDataValidation(dv)

	// set at cell level
	cell3_1.DataValidation = dv

	w, err := os.Create("./testdata/cell_dv_vmerge.xlsx")
	assert.NoError(t, err)
	err = xlsxF.Write(w)
	assert.NoError(t, err)
}

//func TestExistCellOptions(t *testing.T) {
//	f, err := xlsx.OpenFile("./testdata/cell_dv_vmerge.xlsx")
//	assert.NoError(t, err)
//	row, err := f.Sheets[0].Row(0)
//	assert.NoError(t, err)
//	cell := row.GetCell(0)
//	spew.Dump(cell)
//}
