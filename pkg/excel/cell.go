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
	"github.com/tealeg/xlsx/v3"
)

// Cell 单元格
//
// -  A  B  C
// 1  A1 B1 C1
// 2  A2 B2 C2
// 3  A3 B3 C3
type Cell struct {
	// 单元格的值
	Value string
	// 水平合并其他几个单元格
	// 以 A1 为例，默认为 0 表示不合并其他单元格，1 表示合并 A1,B1 两个单元格，2 表示合并 A1,B1,C1 三个单元格
	HorizontalMergeNum int
	// 垂直合并其他几个单元格
	// 以 A1 为例，默认为 0 表示不合并其他单元格，1 表示合并 A1,A2 两个单元格，2 表示合并 A1,A2,A3 三个单元格
	VerticalMergeNum int

	// 单元格的样式
	Style *CellStyle
}

type CellStyle struct {
	IsTitle        bool
	OverwriteStyle *xlsx.Style
}

type CellOption func(*Cell)

func WithStyle(s *CellStyle) CellOption {
	return func(cell *Cell) {
		cell.Style = s
	}
}

func WithIsTitle(isTitle bool) CellOption {
	return func(cell *Cell) {
		if cell.Style == nil {
			cell.Style = &CellStyle{}
		}
		cell.Style.IsTitle = isTitle
	}
}

func WithMergeNum(h, v int) CellOption {
	return func(cell *Cell) {
		cell.HorizontalMergeNum = h
		cell.VerticalMergeNum = v
	}
}

func NewCell(value string, opts ...CellOption) Cell {
	cell := Cell{Value: value}
	for _, opt := range opts {
		opt(&cell)
	}
	return cell
}
func EmptyCell() Cell {
	return Cell{}
}
func EmptyCells(count int) []Cell {
	var cells []Cell
	for i := 0; i < count; i++ {
		cells = append(cells, Cell{})
	}
	return cells
}

// NewHMergeCell 需要在当前行配合 hMergeNum 个 EmptyCell 使用
func NewHMergeCell(value string, hMergeNum int) Cell {
	return NewCell(value, WithMergeNum(hMergeNum, 0))
}

// NewVMergeCell 需要在下方连续 vMergeNum 行配合 EmptyCell 使用；如果下方使用带 Value 的 Cell 也会被 VMergeCell 覆盖，无法展示
func NewVMergeCell(value string, vMergeNum int) Cell {
	return NewCell(value, WithMergeNum(0, vMergeNum))
}
func NewHMergeCellsAuto(value string, hMergeNum int) []Cell {
	return append([]Cell{NewHMergeCell(value, hMergeNum)}, EmptyCells(hMergeNum)...)
}

func NewTitleCell(value string, opts ...CellOption) Cell {
	opts = append(opts, WithIsTitle(true))
	return NewCell(value, opts...)
}

func fulfillCellDataIntoSheet(sheet *xlsx.Sheet, data [][]Cell) {
	for _, cells := range data {
		row := sheet.AddRow()
		for _, cell := range cells {
			xlsxCell := row.AddCell()
			xlsxCell.Value = cell.Value
			xlsxCell.HMerge = cell.HorizontalMergeNum
			xlsxCell.VMerge = cell.VerticalMergeNum
			xlsxCell.SetStyle(defaultStyle())
			if cell.Style != nil {
				if cell.Style.OverwriteStyle != nil {
					xlsxCell.SetStyle(cell.Style.OverwriteStyle)
				} else { // set default
					if cell.Style.IsTitle {
						xlsxCell.SetStyle(defaultTitleCellStyle())
					} else {
						xlsxCell.SetStyle(defaultStyle())
					}
				}
			}
		}
	}

	_ = sheet.ForEachRow(func(r *xlsx.Row) error {
		_ = r.ForEachCell(func(c *xlsx.Cell) error {
			return nil
		}, xlsx.SkipEmptyCells)
		r.SetHeightCM(2)
		return nil
	}, xlsx.SkipEmptyRows)
}

func defaultStyle() *xlsx.Style {
	style := xlsx.NewStyle()
	style.Alignment.Horizontal = "center"
	style.Alignment.Vertical = "center"
	style.Alignment.ShrinkToFit = true
	style.Alignment.WrapText = true
	return style
}

func defaultTitleCellStyle() *xlsx.Style {
	style := defaultStyle()

	style.Border = *xlsx.NewBorder("thin", "thin", "thin", "thin")
	style.Alignment.ShrinkToFit = false
	style.Alignment.WrapText = false

	// set color fill
	fill := *xlsx.NewFill("solid", "FF92D050", "") // set proper color in Excel, and decode to get actual color value
	style.Fill = fill

	return style
}

func convertStringDataToCellData(data [][]string) [][]Cell {
	var cells [][]Cell
	for _, row := range data {
		rowCells := ConvertStringSliceToCellSlice(row)
		cells = append(cells, rowCells)
	}
	return cells
}

func ConvertStringSliceToCellSlice(data []string) []Cell {
	var cells []Cell
	for _, cell := range data {
		cells = append(cells, NewCell(cell))
	}
	return cells
}
