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

package excel

import (
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx/v3"
)

// ExportExcelByCell 支持 cell 粒度配置，可以实现单元格合并，样式调整等
func ExportExcelByCell(w io.Writer, data [][]Cell, sheetName string) error {
	begin := time.Now()
	defer func() {
		end := time.Now()
		logrus.Debugf("export excel cost: %fs", end.Sub(begin).Seconds())
	}()

	file := xlsx.NewFile()
	sheet, err := file.AddSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to add sheet, sheetName: %s, err: %v", sheetName, err)
	}

	for _, cells := range data {
		row := sheet.AddRow()
		for _, cell := range cells {
			xlsxCell := row.AddCell()
			xlsxCell.Value = cell.Value
			xlsxCell.HMerge = cell.HorizontalMergeNum
			xlsxCell.VMerge = cell.VerticalMergeNum
		}
	}

	style := xlsx.NewStyle()
	style.Alignment.Horizontal = "center"
	style.Alignment.Vertical = "center"
	style.Alignment.ShrinkToFit = true
	style.Alignment.WrapText = true

	_ = sheet.ForEachRow(func(r *xlsx.Row) error {
		_ = r.ForEachCell(func(c *xlsx.Cell) error {
			c.SetStyle(style)
			return nil
		}, xlsx.SkipEmptyCells)
		r.SetHeightCM(1.5)
		return nil
	}, xlsx.SkipEmptyRows)

	return write(w, file, sheetName)
}
