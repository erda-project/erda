package sqllint

import (
	"bytes"
	"strconv"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type IndexLengthLinter struct {
	script Script
	err    error
	text   string
}

func NewIndexLengthLinter(script Script) Rule {
	return &IndexLengthLinter{script: script}
}

func (l *IndexLengthLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}
	var colNames = make(map[string]int, 0)
	for _, col := range stmt.Cols {
		colName := ddlconv.ExtractColName(col)
		if colName == "" {
			continue
		}
		if col.Tp == nil {
			continue
		}
		colNames[colName] = col.Tp.Flen
	}
	for _, c := range stmt.Constraints {
		var length, firstLen int
		for _, key := range c.Keys {
			if key.Length > 0 {
				length += key.Length
				firstLen = key.Length
				continue
			}
			if key.Column == nil {
				continue
			}
			if l, ok := colNames[key.Column.Name.String()]; ok {
				length += l
				firstLen = l
				continue
			}
		}
		if len(c.Keys) == 1 && length*4 > 767 {
			l.err = NewLintError(l.script, l.text, "索引长度错误: 单列索引长度不得 > 767",
				func(line []byte) bool {
					firstLenS := strconv.FormatInt(int64(firstLen), 10)
					return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(firstLenS)))
				})
			return in, true
		}
		if len(c.Keys) > 1 && length*4 > 3072 {
			_, num := getLintLine(l.script.Data(), []byte(l.text), func(_ []byte) bool {
				return false
			})
			l.err = LintError{
				Stmt:   l.text,
				Lint:   "索引长度错误: 联合索引长度不得 > 3072",
				LintNo: num,
			}
			return in, true
		}
	}

	return in, true
}

func (l *IndexLengthLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IndexLengthLinter) Error() error {
	return l.err
}
