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

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_baseinfo"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_customfield"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_issue"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_iteration"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_label"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_state"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets/sheet_user"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

func ExportFile(w io.Writer, data *vars.DataForFulfill) (err error) {
	xlsxFile := excel.NewFile()
	multiWriter := io.MultiWriter(w)
	defer func() {
		if err == nil {
			err = excel.WriteFile(multiWriter, xlsxFile, data.ExportOnly.FileNameWithExt)
		}
	}()

	// issue sheet
	issueExcelRows, err := ExportIssueSheetLines(data)
	if err != nil {
		return fmt.Errorf("failed to export issue sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(issueExcelRows), vars.NameOfSheetIssue); err != nil {
		return fmt.Errorf("failed to add issue sheet, err: %v", err)
	}

	// only full export need to export other sheets
	if !data.IsFullExport() {
		return nil
	}

	// user sheet
	userExcelRows, err := sheet_user.GenUserSheet(data)
	if err != nil {
		return fmt.Errorf("failed to gen user sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(userExcelRows), vars.NameOfSheetUser); err != nil {
		return fmt.Errorf("failed to add user sheet, err: %v", err)
	}
	// label sheet
	labelExcelRows, err := sheet_label.GenLabelSheet(data)
	if err != nil {
		return fmt.Errorf("failed to gen label sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(labelExcelRows), vars.NameOfSheetLabel); err != nil {
		return fmt.Errorf("failed to add label sheet, err: %v", err)
	}
	// custom field sheet
	customFieldExcelRows, err := sheet_customfield.GenCustomFieldSheet(data)
	if err != nil {
		return fmt.Errorf("failed to gen custom field sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(customFieldExcelRows), vars.NameOfSheetCustomField); err != nil {
		return fmt.Errorf("failed to add custom field sheet, err: %v", err)
	}
	// meta sheet
	baseInfoExcelRows, err := sheet_baseinfo.GenBaseInfoSheet(data)
	if err != nil {
		return fmt.Errorf("failed to gen meta sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(baseInfoExcelRows), vars.NameOfSheetBaseInfo); err != nil {
		return fmt.Errorf("failed to add meta sheet, err: %v", err)
	}
	// iteration sheet
	iterationExcelRows, err := sheet_iteration.GenIterationSheet(data)
	if err != nil {
		return fmt.Errorf("failed to gen iteration sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(iterationExcelRows), vars.NameOfSheetIteration); err != nil {
		return fmt.Errorf("failed to add iteration sheet, err: %v", err)
	}
	// state sheet
	stateExcelRows, err := sheet_state.GenStateSheet(data)
	if err != nil {
		return fmt.Errorf("failed to gen state sheet, err: %v", err)
	}
	if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(stateExcelRows), vars.NameOfSheetState); err != nil {
		return fmt.Errorf("failed to add state sheet, err: %v", err)
	}

	return nil
}

func ExportIssueSheetLines(data *vars.DataForFulfill) (excel.Rows, error) {
	mapByColumns, err := sheet_issue.GenIssueSheetTitleAndDataByColumn(data)
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
