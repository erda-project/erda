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

package sqllint

import (
	"github.com/erda-project/erda/pkg/sqllint/linters"
	"github.com/erda-project/erda/pkg/sqllint/rules"
)

// DefaultRulers is default DefaultRulers
var DefaultRulers = map[string]rules.Ruler{
	"BooleanFieldLinter":          linters.NewBooleanFieldLinter,
	"CharsetLinter":               linters.NewCharsetLinter,
	"ColumnNameLinter":            linters.NewColumnNameLinter,
	"ColumnCommentLinter":         linters.NewColumnCommentLinter,
	"DDLDMLLinter":                linters.NewDDLDMLLinter,
	"DestructLinter":              linters.NewDestructLinter,
	"FloatDoubleLinter":           linters.NewFloatDoubleLinter,
	"ForeignKeyLinter":            linters.NewForeignKeyLinter,
	"IndexLengthLinter":           linters.NewIndexLengthLinter,
	"IndexNameLinter":             linters.NewIndexNameLinter,
	"KeywordsLinter":              linters.NewKeywordsLinter,
	"IDExistsLinter":              linters.NewIDExistsLinter,
	"IDTypeLinter":                linters.NewIDTypeLinter,
	"IDIsPrimaryLinter":           linters.NewIDIsPrimaryLinter,
	"CreatedAtExistsLinter":       linters.NewCreatedAtExistsLinter,
	"CreatedAtTypeLinter":         linters.NewCreatedAtTypeLinter,
	"CreatedAtDefaultValueLinter": linters.NewCreatedAtDefaultValueLinter,
	"UpdatedAtExistsLinter":       linters.NewUpdatedAtExistsLinter,
	"UpdatedAtTypeLinter":         linters.NewUpdatedAtTypeLinter,
	"UpdatedAtDefaultValueLinter": linters.NewUpdatedAtDefaultValueLinter,
	"UpdatedAtOnUpdateLinter":     linters.NewUpdatedAtOnUpdateLinter,
	"NotNullLinter":               linters.NewNotNullLinter,
	"TableCommentLinter":          linters.NewTableCommentLinter,
	"TableNameLinter":             linters.NewTableNameLinter,
	"VarcharLengthLinter":         linters.NewVarcharLengthLinter,
}
