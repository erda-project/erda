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

package issueexcel

import (
	"fmt"
	"io"
	"os"

	"github.com/erda-project/erda/pkg/excel"
)

func ExportFile(w io.Writer, data DataForFulfill) error {
	xlsxFile := excel.NewXLSXFile()
	// issue sheet
	issueExcelRows, err := ExportIssueSheetLines(data)
	if err != nil {
		return fmt.Errorf("failed to export issue sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(issueExcelRows), "issue"); err != nil {
		return fmt.Errorf("failed to add issue sheet, err: %v", err)
	}
	// user sheet
	userExcelRows, err := data.genUserSheet()
	if err != nil {
		return fmt.Errorf("failed to gen user sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(userExcelRows), "user"); err != nil {
		return fmt.Errorf("failed to add user sheet, err: %v", err)
	}
	// label sheet
	labelExcelRows, err := data.genLabelSheet()
	if err != nil {
		return fmt.Errorf("failed to gen label sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(labelExcelRows), "label"); err != nil {
		return fmt.Errorf("failed to add label sheet, err: %v", err)
	}
	// custom field sheet
	customFieldExcelRows, err := data.genCustomFieldSheet()
	if err != nil {
		return fmt.Errorf("failed to gen custom field sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(customFieldExcelRows), "custom_field"); err != nil {
		return fmt.Errorf("failed to add custom field sheet, err: %v", err)
	}
	// export file
	f, err := os.OpenFile("./gen2.xlsx", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer f.Close()
	multiWriter := io.MultiWriter(w, f)
	return excel.WriteFile(multiWriter, xlsxFile, data.FileNameWithExt)
}

func ExportIssueSheetLines(data DataForFulfill) (excel.Rows, error) {
	mapByColumns, err := data.genIssueSheetTitleAndDataByColumn()
	if err != nil {
		return nil, fmt.Errorf("failed to gen sheet title and data by column, err: %v", err)
	}
	excelRows, err := mapByColumns.ConvertToExcelSheet()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to excel sheet, err: %v", err)
	}
	return excelRows, nil
}

func convertExcelRowsToCells(rows excel.Rows) [][]excel.Cell {
	cells := make([][]excel.Cell, len(rows))
	for i, row := range rows {
		cells[i] = make([]excel.Cell, len(row))
		for j, cell := range row {
			cells[i][j] = cell
		}
	}
	return cells
}
