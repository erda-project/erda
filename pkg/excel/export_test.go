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

	"github.com/davecgh/go-spew/spew"
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

func TestSheetValidation(t *testing.T) {
	xlsxF := xlsx.NewFile()
	sheet, err := xlsxF.AddSheet("test")
	assert.NoError(t, err)
	row := sheet.AddRow()
	cell1 := row.AddCell()
	cell2 := row.AddCell()
	dv1 := xlsx.NewDataValidation(0, 0, 1, 1, false)
	//dv2 := xlsx.NewDataValidation(0, 1, 0, 1, false)
	err = dv1.SetDropList([]string{"待处理", "进行中", "已完成"})
	assert.NoError(t, err)
	//err = dv2.SetDropList([]string{"需求", "任务", "缺陷"})
	//assert.NoError(t, err)
	fill := *xlsx.NewFill("solid", "FF92D050", "")
	style := xlsx.NewStyle()
	style.Fill = fill
	cell1.SetStyle(style)
	cell2.SetStyle(style)
	sheet.AddDataValidation(dv1)

	writeF, err := os.Create("./testdata/cell.xlsx")
	err = xlsxF.Write(writeF)
	assert.NoError(t, err)
}

func TestExistCellOptions(t *testing.T) {
	f, err := xlsx.OpenFile("./testdata/cell.xlsx")
	assert.NoError(t, err)
	row, err := f.Sheets[0].Row(0)
	assert.NoError(t, err)
	cell := row.GetCell(0)
	spew.Dump(cell)
}
