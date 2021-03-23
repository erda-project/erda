package sqllint

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type FloatDoubleLinter struct {
	script Script
	err    error
	text   string
}

func NewFloatDoubleLinter(script Script) Rule {
	return &FloatDoubleLinter{script: script}
}

func (l *FloatDoubleLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	t := ddlconv.ExtractColType(col)
	t = strings.ToLower(t)
	if strings.Contains(t, "float") || strings.Contains(t, "double") {
		l.err = NewLintError(l.script, l.text, "字段类型错误: 小数类型应当使用 decimal, 而不是 float、double",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("float")) ||
					bytes.Contains(bytes.ToLower(line), []byte("double"))
			})
	}

	return in, false
}

func (l *FloatDoubleLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *FloatDoubleLinter) Error() error {
	return l.err
}
