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
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
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
	generateCell(sheet, data)
	return write(w, file, sheetName)
}

func generateCell(sheet *xlsx.Sheet, data [][]Cell) {
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
}

func WriteExcelBuffer(data [][]Cell, sheetName string) (*bytes.Buffer, error) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to add sheet, sheetName: %s, err: %v", sheetName, err)
	}
	generateCell(sheet, data)
	var buff bytes.Buffer
	if err := file.Write(&buff); err != nil {
		return nil, errors.Errorf("failed to write content, sheetName: %s, err: %v", sheetName, err)
	}
	return &buff, nil
}

type XlsxFile struct {
	file *xlsx.File
}

func NewXLSXFile() *XlsxFile {
	return &XlsxFile{file: xlsx.NewFile()}
}

func AddSheetByCell(f *XlsxFile, data [][]Cell, sheetName string) error {
	sheet, err := f.file.AddSheet(sheetName)
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
	return nil
}

func WriteFile(w io.Writer, f *XlsxFile, filename string) error {
	return write(w, f.file, filename)
}
