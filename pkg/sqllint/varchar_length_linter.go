package sqllint

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"
)

type VarcharLengthLinter struct {
	script Script
	err    error
	text   string
}

func NewVarcharLengthLinter(script Script) Rule {
	return &VarcharLengthLinter{script: script}
}

func (l *VarcharLengthLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	if col.Tp != nil &&
		strings.Contains(strings.ToLower(col.Tp.String()), "varchar") &&
		col.Tp.Flen > 5000 {
		l.err = NewLintError(l.script, l.text, "字段类型错误: varchar 类型长度不可 > 5000",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte(col.Tp.String()))

			})
	}

	return in, true
}

func (l *VarcharLengthLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *VarcharLengthLinter) Error() error {
	return l.err
}
