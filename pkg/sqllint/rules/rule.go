package rules

import (
	"github.com/erda-project/erda/pkg/sqllint/script"
	"github.com/pingcap/parser/ast"
)

// Rule is an Error and SQL ast visitor,
// can accept a SQL stmt and lint it.
type Rule interface {
	ast.Visitor

	Error() error
}

// Ruler is a function that returns a Rule interface
type Ruler func(script script.Script) Rule
