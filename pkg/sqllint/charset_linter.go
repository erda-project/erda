package sqllint

import (
	"strings"

	"github.com/pingcap/parser/ast"
)

const UTF8MB4 = "utf8mb4"

type CharsetLinter struct {
	script Script
	err    error
	text   string
}

func NewCharsetLinter(script Script) Rule {
	return &CharsetLinter{script: script}
}

func (l *CharsetLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, opt := range stmt.Options {
		if opt.Tp == ast.TableOptionCharset && strings.ToLower(opt.StrValue) == UTF8MB4 {
			return in, true
		}
	}
	l.err = NewLintError(l.script, l.text, "表字符集错误: 应当显示声明为 CHARSET = utf8mb4",
		func(line []byte) bool {
			return false
		})
	return in, true
}

func (l *CharsetLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *CharsetLinter) Error() error {
	return l.err
}
