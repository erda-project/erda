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

package configuration

import (
	"github.com/erda-project/erda/pkg/database/sqllint/linters"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
)

const (
	defaultAllowedDDLs = linters.CreateTableStmt | linters.CreateIndexStmt | linters.DropIndexStmt |
		linters.AlterTableOption | linters.AlterTableAddColumns | linters.AlterTableAddConstraint |
		linters.AlterTableDropIndex | linters.AlterTableModifyColumn | linters.AlterTableChangeColumn |
		linters.AlterTableAlterColumn | linters.AlterTableRenameIndex
	defaultAllowedDMLs = linters.SelectStmt | linters.UnionStmt | linters.InsertStmt |
		linters.UpdateStmt | linters.ShowStmt
)

// DefaultRulers returns the default SQL lint rulers
func DefaultRulers() []rules.Ruler {
	return defaultRulers
}

var defaultRulers = []rules.Ruler{
	linters.NewAllowedStmtLinter(defaultAllowedDDLs, defaultAllowedDMLs),
	linters.NewBooleanFieldLinter,
	linters.NewCharsetLinter,
	linters.NewColumnCommentLinter,
	linters.NewColumnNameLinter,

	linters.NewCreatedAtDefaultValueLinter,
	linters.NewCreatedAtExistsLinter,
	linters.NewCreatedAtTypeLinter,

	linters.NewDestructLinter,
	linters.NewFloatDoubleLinter,
	linters.NewForeignKeyLinter,

	linters.NewIDExistsLinter,
	linters.NewIDIsPrimaryLinter,
	linters.NewIDTypeLinter,

	linters.NewIndexLengthLinter,
	linters.NewIndexNameLinter,

	linters.NewKeywordsLinter,
	linters.NewNotNullLinter,
	linters.NewTableCommentLinter,
	linters.NewTableNameLinter,

	linters.NewUpdatedAtExistsLinter,
	linters.NewUpdatedAtTypeLinter,
	linters.NewUpdatedAtDefaultValueLinter,
	linters.NewUpdatedAtOnUpdateLinter,

	linters.NewVarcharLengthLinter,
	linters.NewCompleteInsertLinter,
	linters.NewManualTimeSetterLinter,

	linters.NewExplicitCollationLinter,
}
