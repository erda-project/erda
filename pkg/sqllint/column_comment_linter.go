package sqllint

import (
	"bytes"

	"github.com/pingcap/parser/ast"
)

type ColumnCommentLinter struct {
	script Script
	err    error
	text   string
}

func NewColumnCommentLinter(script Script) Rule {
	return &ColumnCommentLinter{script: script}
}

func (l *ColumnCommentLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	for _, opt := range col.Options {
		if opt.Tp == ast.ColumnOptionComment {
			return in, true
		}
	}
	l.err = NewLintError(l.script, l.text, "字段定义子句缺少必要的 option: 应当显示声明 comment",
		func(line []byte) bool {
			if col.Name == nil {
				return true
			}
			return bytes.Contains(line, bytes.ToLower([]byte(col.Name.String())))
		})
	return in, true
}

func (l *ColumnCommentLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *ColumnCommentLinter) Error() error {
	return l.err
}
