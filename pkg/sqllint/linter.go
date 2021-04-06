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
	"github.com/pingcap/parser/ast"
)

const (
	ID        = "id"
	CreatedAt = "created_at"
	UpdatedAt = "updated_at"
)

var Rules = map[string]NewRule{
	"BooleanFieldLinter":          NewBooleanFieldLinter,
	"CharsetLinter":               NewCharsetLinter,
	"ColumnNameLinter":            NewColumnNameLinter,
	"ColumnCommentLinter":         NewColumnCommentLinter,
	"DDLDMLLinter":                NewDDLDMLLinter,
	"DestructLinter":              NewDestructLinter,
	"FloatDoubleLinter":           NewFloatDoubleLinter,
	"ForeignKeyLinter":            NewForeignKeyLinter,
	"IndexLengthLinter":           NewIndexLengthLinter,
	"IndexNameLinter":             NewIndexNameLinter,
	"KeywordsLinter":              NewKeywordsLinter,
	"IDExistsLinter":              NewIDExistsLinter,
	"IDTypeLinter":                NewIDTypeLinter,
	"IDIsPrimaryLinter":           NewIDIsPrimaryLinter,
	"CreatedAtExistsLinter":       NewCreatedAtExistsLinter,
	"CreatedAtTypeLinter":         NewCreatedAtTypeLinter,
	"CreatedAtDefaultValueLinter": NewCreatedAtDefaultValueLinter,
	"UpdatedAtExistsLinter":       NewUpdatedAtExistsLinter,
	"UpdatedAtTypeLinter":         NewUpdatedAtTypeLinter,
	"UpdatedAtDefaultValueLinter": NewUpdatedAtDefaultValueLinter,
	"UpdatedAtOnUpdateLinter":     NewUpdatedAtOnUpdateLinter,
	"NotNullLinter":               NewNotNullLinter,
	"TableCommentLinter":          NewTableCommentLinter,
	"TableNameLinter":             NewTableNameLinter,
	"VarcharLengthLinter":         NewVarcharLengthLinter,
}

type Rule interface {
	ast.Visitor

	Error() error
}

type NewRule func(script Script) Rule
