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
