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

package sheet_customfield

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/excel"
)

func (h *Handler) ExportSheet(data *vars.DataForFulfill) (excel.Rows, error) {
	var lines excel.Rows
	// title: custom field id, custom field name, custom field type, custom field value
	title := excel.Row{
		excel.NewTitleCell("custom field id"),
		excel.NewTitleCell("custom field name"),
		excel.NewTitleCell("custom field type"),
		excel.NewTitleCell("custom field detail (json)"),
	}
	lines = append(lines, title)
	// data
	for propertyType, properties := range data.CustomFieldMapByTypeName {
		for _, cf := range properties {
			cfInfo, err := json.Marshal(cf)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal custom field info, custom field id: %d, err: %v", cf.PropertyID, err)
			}
			lines = append(lines, excel.Row{
				excel.NewCell(strconv.FormatInt(cf.PropertyID, 10)),
				excel.NewCell(cf.PropertyName),
				excel.NewCell(propertyType.String()),
				excel.NewCell(string(cfInfo)),
			})
		}
	}

	return lines, nil
}
