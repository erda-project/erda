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

func ImportFile(r io.Reader, data *vars.DataForFulfill) error {
	// decode to sheets
	df, err := excel.DecodeToSheets(r)
	if err != nil {
		return fmt.Errorf("failed to decode excel, err: %v", err)
	}
	data.ImportOnly.DecodedFile = df
	// compatible
	data.JudgeIfIsOldExcelFormat(df)

	handlers := []sheets.Importer{
		&sheet_baseinfo.Handler{},
		&sheet_user.Handler{},
		&sheet_label.Handler{},
		&sheet_customfield.Handler{},
		&sheet_iteration.Handler{},
		&sheet_state.Handler{},
		&sheet_issue.Handler{},
	}

	// 1. decode sheets
	for _, h := range handlers {
		// check sheet exist or not
		sheet, ok := df.Sheets.M[h.SheetName()]
		if !ok {
			continue
		}
		if err := h.DecodeSheet(data, sheet); err != nil {
			return fmt.Errorf("failed to docode sheet %q, err: %v", h.SheetName(), err)
		}
	}

	if checkImportError(data, df, handlers) == signalStop {
		return nil
	}

	// 2. before create issues
	// - add member to project
	// - create label
	// - create custom-field
	// - create iteration
	// - create state
	// - check issues (relation, user, iteration, state, ...)
	for _, h := range handlers {
		// all handle can do sth before create issue, even if sheet not exist
		if err := h.BeforeCreateIssues(data); err != nil {
			return fmt.Errorf("failed to do before create issue, sheet: %q, err: %v", h.SheetName(), err)
		}
	}

	// 3. create issues
	for _, h := range handlers {
		hh, ok := h.(sheets.ImporterCreateIssues)
		if !ok {
			continue
		}
		if err := hh.CreateIssues(data); err != nil {
			return fmt.Errorf("failed to create issue, sheet: %q, err: %v", h.SheetName(), err)
		}
	}

	// 4. after create issues
	// - create relation: label <-> issue
	// - create relation: custom-field <-> issue
	// - create relation: issue <-> issue
	for _, h := range handlers {
		// all handle can do sth after create issue, even if sheet not exist
		if err := h.AfterCreateIssues(data); err != nil {
			return fmt.Errorf("failed to do after create issue, sheet: %q, err: %v", h.SheetName(), err)
		}
	}

	return nil
}

type signal int

const (
	signalContinue signal = iota
	signalStop
)

func checkImportError(data *vars.DataForFulfill, df *excel.DecodedFile, handlers []sheets.Importer) signal {
	if len(data.ImportOnly.Errs) == 0 {
		return signalContinue
	}
	for _, h := range handlers {
		hh, ok := h.(sheets.ImporterAppendErrorColumn)
		if !ok {
			continue
		}
		sheet := df.Sheets.M[h.SheetName()]
		hh.AppendErrorColumn(data, sheet)
	}
	return signalStop
}
