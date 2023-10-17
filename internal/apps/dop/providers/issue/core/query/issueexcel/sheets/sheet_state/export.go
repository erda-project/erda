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

package sheet_state

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

func (h *Handler) ExportSheet(data *vars.DataForFulfill) (*sheets.RowsForExport, error) {
	var lines excel.Rows

	// title: state (JSON), state_relation (JSON)
	title := excel.Row{
		excel.NewTitleCell("state (json)"),
		excel.NewTitleCell("state_relation (json)"),
	}
	lines = append(lines, title)

	// data
	stateBytes, err := json.Marshal(data.ExportOnly.States)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state info, err: %v", err)
	}
	stateRelationBytes, err := json.Marshal(data.ExportOnly.StateRelations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state relation info, err: %v", err)
	}
	lines = append(lines, excel.Row{
		excel.NewCell(string(stateBytes)),
		excel.NewCell(string(stateRelationBytes)),
	})

	return sheets.NewRowsForExport(lines), nil
}
