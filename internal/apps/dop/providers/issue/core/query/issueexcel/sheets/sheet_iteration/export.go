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

package sheet_iteration

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

func (h *Handler) ExportSheet(data *vars.DataForFulfill) (excel.Rows, error) {
	// if AllProjectIssues=true, then export all iterations
	// otherwise, just export iterations related to issues
	relatedIterationMapByID := make(map[int64]struct{})
	if !data.ExportOnly.AllProjectIssues {
		for _, issue := range data.ExportOnly.Issues {
			relatedIterationMapByID[issue.IterationID] = struct{}{}
		}
	}

	var lines excel.Rows
	// title: iteration id, iteration name, iteration info (JSON)
	title := excel.Row{
		excel.NewTitleCell("iteration id"),
		excel.NewTitleCell("iteration name"),
		excel.NewTitleCell("iteration detail (json)"),
	}
	lines = append(lines, title)
	// data
	for _, iteration := range data.IterationMapByID {
		iteration := iteration
		if iteration.ID <= 0 {
			continue
		}
		if !data.ExportOnly.AllProjectIssues {
			// only related iteration need to be exported
			if _, ok := relatedIterationMapByID[int64(iteration.ID)]; !ok {
				continue
			}
		}
		b, err := json.Marshal(iteration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal iteration info, iteration id: %d, err: %v", iteration.ID, err)
		}
		lines = append(lines, excel.Row{
			excel.NewCell(strconv.FormatUint(iteration.ID, 10)),
			excel.NewCell(iteration.Title),
			excel.NewCell(string(b)),
		})
	}

	return lines, nil
}
