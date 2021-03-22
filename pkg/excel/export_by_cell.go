package excel

import (
	"fmt"
	"io"

	"github.com/tealeg/xlsx/v3"
)

// ExportExcelByCell 支持 cell 粒度配置，可以实现单元格合并，样式调整等
func ExportExcelByCell(w io.Writer, data [][]Cell, sheetName string) error {
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
