package sqllint

import (
	"bytes"

	"github.com/pingcap/parser/ast"
)

type ForeignKeyLinter struct {
	script Script
	err    error
	text   string
}

func NewForeignKeyLinter(script Script) Rule {
	return &ForeignKeyLinter{script: script}
}

func (l *ForeignKeyLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	constraint, ok := in.(*ast.Constraint)
	if !ok {
		return in, false
	}

	if constraint.Tp == ast.ConstraintForeignKey {
		l.err = NewLintError(l.script, l.text, "使用了外键: 一切外键概念必须在应用层表达",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("foreign"))
			})
	}

	return in, false
}

func (l *ForeignKeyLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *ForeignKeyLinter) Error() error {
	return l.err
}
