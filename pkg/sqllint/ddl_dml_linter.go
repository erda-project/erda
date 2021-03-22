package sqllint

import (
	"github.com/pingcap/parser/ast"
)

type DDLDMLLinter struct {
	script Script
	err    error
	text   string
}

func NewDDLDMLLinter(script Script) Rule {
	return &DDLDMLLinter{script: script}
}

func (l *DDLDMLLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case ast.DDLNode, ast.DMLNode:
	default:
		l.err = NewLintError(l.script, l.text,
			"语言类型错误: 只能包含数据定义语言(DDL)、数据操作语言(DML), 不可以包含数据库操作语言(DCL)、事务控制语言(TCL)",
			func(line []byte) bool {
				return true
			})
	}

	return in, true
}

func (l *DDLDMLLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, false
}

func (l *DDLDMLLinter) Error() error {
	return l.err
}
