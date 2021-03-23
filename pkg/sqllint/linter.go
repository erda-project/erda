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
