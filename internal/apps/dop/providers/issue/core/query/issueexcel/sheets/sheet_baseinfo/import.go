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

package sheet_baseinfo

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

const (
	cellMeta = "meta"
)

type Handler struct{ sheets.DefaultImporter }

func (h *Handler) SheetName() string { return vars.NameOfSheetBaseInfo }

func (h *Handler) DecodeSheet(data *vars.DataForFulfill, s *excel.Sheet) error {
	if data.IsOldExcelFormat() {
		return nil
	}
	sheet := s.UnmergedSlice
	if len(sheet) != 1 {
		return fmt.Errorf("invalid base info sheet, rows: %d", len(sheet))
	}
	if len(sheet[0]) != 2 {
		return fmt.Errorf("invalid base info sheet, cols: %d", len(sheet[0]))
	}
	if sheet[0][0] != cellMeta {
		return fmt.Errorf("invalid base info sheet, first col: %s", sheet[0][0])
	}
	var meta vars.DataForFulfillImportOnlyBaseInfo
	if err := json.Unmarshal([]byte(sheet[0][1]), &meta); err != nil {
		return fmt.Errorf("failed to unmarshal meta, err: %v", err)
	}
	// set into data
	data.ImportOnly.Sheets.Optional.BaseInfo = &meta
	return nil
}
