package sqllint

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/swagger/openapi/ddl"
)

type TableNameLinter struct {
	script Script
	err    error
	text   string
}

func NewTableNameLinter(script Script) Rule {
	return &TableNameLinter{script: script}
}

func (l *TableNameLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true // 只有 create stmt 才验证表名
	}

	name := ddl.ExtractCreateName(stmt)
	if name == "" {
		return in, true
	}

	if compile := regexp.MustCompile(`^[0-9a-z_]{1,}$`); !compile.Match([]byte(name)) {
		l.err = NewLintError(l.script, l.text, "表名不合法, 只能包含小写英文字母、数字、下划线",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
			})
		return in, true
	}

	if w := name[0]; '0' <= w && w <= '9' {
		l.err = NewLintError(l.script, l.text, "表名不合法，不能以数字开头",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
			})
		return in, true
	}

	words := strings.Split(name, "_")
	for _, w := range words {
		if _, err := strconv.ParseInt(w, 10, 64); w == "" || err == nil {
			l.err = NewLintError(l.script, l.text, "表名不合法, 两下划线中至少包含一个字母",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
				})
			return in, true
		}
	}

	return in, true
}

func (l *TableNameLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *TableNameLinter) Error() error {
	return l.err
}
