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

package sheet_user

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/sheets"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query/issueexcel/vars"
	"github.com/erda-project/erda/pkg/desensitize"
	"github.com/erda-project/erda/pkg/excel"
)

func (h *Handler) ExportSheet(data *vars.DataForFulfill) (*sheets.RowsForExport, error) {
	var lines excel.Rows
	// title: user id, user name, user info (JSON)
	title := excel.Row{
		excel.NewTitleCell("user id"),
		excel.NewTitleCell("user name"),
		excel.NewTitleCell("user detail (json)"),
	}
	lines = append(lines, title)
	// data
	for _, user := range data.ProjectMemberByUserID {
		// desensitize data
		user.Name = desensitize.Name(user.Name)
		user.Nick = desensitize.Name(user.Nick)
		user.Mobile = desensitize.Mobile(user.Mobile)
		user.Email = desensitize.Email(user.Email)

		userInfo, err := json.Marshal(user)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal user info, user id: %s, err: %v", user.UserID, err)
		}
		lines = append(lines, excel.Row{
			excel.NewCell(user.UserID),
			excel.NewCell(user.Nick),
			excel.NewCell(string(userInfo)),
		})
	}

	return sheets.NewRowsForExport(lines), nil
}
