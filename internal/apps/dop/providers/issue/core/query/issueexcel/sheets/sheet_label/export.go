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

package sheet_label

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

func (h *Handler) ExportSheet(data *vars.DataForFulfill) (*sheets.RowsForExport, error) {
	var lines excel.Rows
	// title: label id, label name, label detail (JSON)
	title := excel.Row{
		excel.NewTitleCell("label id"),
		excel.NewTitleCell("label name"),
		excel.NewTitleCell("label detail (json)"),
	}
	lines = append(lines, title)
	// data
	// collect labels from issues
	labelMap := make(map[int64]*pb.ProjectLabel)
	for _, issue := range data.ExportOnly.Issues {
		for _, label := range issue.LabelDetails {
			if _, ok := labelMap[label.Id]; !ok {
				labelMap[label.Id] = label
			}
		}
	}
	for _, label := range labelMap {
		labelInfo, err := json.Marshal(label)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal label info, label id: %d, err: %v", label.Id, err)
		}
		lines = append(lines, excel.Row{
			excel.NewCell(strconv.FormatInt(label.Id, 10)),
			excel.NewCell(label.Name),
			excel.NewCell(string(labelInfo)),
		})
	}

	return sheets.NewRowsForExport(lines), nil
}
