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

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
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

	handlers := []sheets.Exporter{
		&sheet_baseinfo.Handler{},
		&sheet_issue.Handler{},
		&sheet_user.Handler{},
		&sheet_label.Handler{},
		&sheet_customfield.Handler{},
		&sheet_iteration.Handler{},
		&sheet_state.Handler{},
	}

	for _, h := range handlers {
		// only full export need to export other sheets
		if !data.IsFullExport() && h.SheetName() != vars.NameOfSheetIssue {
			continue
		}

		rows, err := h.ExportSheet(data)
		if err != nil {
			return fmt.Errorf("failed to gen sheet %q, err: %v", h.SheetName(), err)
		}
		if err := excel.AddSheetByCell(xlsxFile, convertExcelRowsToCells(rows), h.SheetName()); err != nil {
			return fmt.Errorf("failed to add issue %q, err: %v", h.SheetName(), err)
		}
	}

	return nil
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
