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
	// compatible
	data.JudgeIfIsOldExcelFormat(df)

	handlers := []sheets.Importer{
		&sheet_baseinfo.Handler{},
		&sheet_issue.Handler{},
		&sheet_user.Handler{},
		&sheet_label.Handler{},
		&sheet_customfield.Handler{},
		&sheet_iteration.Handler{},
		&sheet_state.Handler{},
	}

	for _, h := range handlers {
		if err := h.ImportSheet(data, df); err != nil {
			return fmt.Errorf("failed to decode sheet %q, err: %v", h.SheetName(), err)
		}
	}

	// 先创建或更新所有 issues，再创建或更新所有关联关系

	// 创建或更新 issues
	// 更新 model 里的相关关联 ID 字段，比如 L1 转换为具体的 ID
	issues, issueModelMapByIssueID, err := sheet_issue.CreateOrUpdateIssues(data, data.ImportOnly.Sheets.Must.IssueInfo)
	if err != nil {
		return fmt.Errorf("failed to create or update issues, err: %v", err)
	}

	// 先将数据进行合并，以 label 为例:
	// - 收集 issue 里的 label
	// - 与 label sheet 里的 label 进行合并
	// - 创建或更新 label
	// - 创建或更新关联 issue 与 label 的关联关系

	// create issue label
	if err := sheet_label.CreateLabelFromIssueSheet(data); err != nil {
		return fmt.Errorf("failed to create label from issue sheet, err: %v", err)
	}

	// create label relation
	if err := sheet_label.CreateIssueLabelRelations(data, issues, issueModelMapByIssueID); err != nil {
		return fmt.Errorf("failed to create issue label relations, err: %v", err)
	}
	// create custom field relation
	if err := sheet_customfield.CreateIssueCustomFieldRelation(data, issues, issueModelMapByIssueID); err != nil {
		return fmt.Errorf("failed to create issue custom field relations, err: %v", err)
	}
	// create issue relation
	if err := sheet_issue.CreateIssueRelations(data, issues, issueModelMapByIssueID); err != nil {
		return fmt.Errorf("failed to create issue relations, err: %v", err)
	}
	return nil
}
