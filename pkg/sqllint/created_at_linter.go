package sqllint

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/tidb/types"
)

type CreatedAtExistsLinter struct {
	script Script
	err    error
	text   string
}

func NewCreatedAtExistsLinter(script Script) Rule {
	return &CreatedAtExistsLinter{script: script}
}

func (l *CreatedAtExistsLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// 查找是否存在名为 created_at 的键
	for _, col := range createStmt.Cols {
		if col.Name != nil && strings.ToLower(col.Name.String()) == CreatedAt {
			return in, true
		}
	}

	// 如果没有 created_at 字段
	l.err = NewLintError(l.script, l.text, "缺少必要字段: created_at", func(line []byte) bool {
		return false
	})

	return in, true
}

func (l *CreatedAtExistsLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *CreatedAtExistsLinter) Error() error {
	return l.err
}

type CreatedAtTypeLinter struct {
	script Script
	err    error
	text   string
}

func NewCreatedAtTypeLinter(script Script) Rule {
	return &CreatedAtTypeLinter{script: script}
}

func (l *CreatedAtTypeLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, col := range createStmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != CreatedAt {
			continue
		}

		// 存在名为 created_at 的键, 则检查类型
		if strings.ToLower(col.Tp.String()) == types.DateTimeStr {
			return in, true
		}

		l.err = NewLintError(l.script, l.text, "created_at 类型错误: 应当为 datetime", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte("created_at"))
		})

		return in, true
	}

	return in, true
}

func (l *CreatedAtTypeLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *CreatedAtTypeLinter) Error() error {
	return l.err
}

type CreatedAtDefaultValueLinter struct {
	script Script
	err    error
	text   string
}

func NewCreatedAtDefaultValueLinter(script Script) Rule {
	return &CreatedAtDefaultValueLinter{script: script}
}

func (l *CreatedAtDefaultValueLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, col := range createStmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != CreatedAt {
			continue
		}

		// 检查是否自动跟踪创建时间
		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionDefaultValue {
				if expr, ok := opt.Expr.(*ast.FuncCallExpr); ok && strings.ToUpper(expr.FnName.String()) == "CURRENT_TIMESTAMP" {
					return in, true
				}
			}
		}

		l.err = NewLintError(l.script, l.text, "created_at DEFAULT VALUE 错误: 应当为 CURRENT_TIMESTAMP",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("created_at"))
			})

		return in, true
	}

	return in, true
}

func (l *CreatedAtDefaultValueLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *CreatedAtDefaultValueLinter) Error() error {
	return l.err
}
