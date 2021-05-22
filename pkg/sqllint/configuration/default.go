// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package configuration

import (
	"github.com/erda-project/erda/pkg/sqllint/linters"
	"github.com/erda-project/erda/pkg/sqllint/rules"
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
}
