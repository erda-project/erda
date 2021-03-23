package sqllint

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/tidb/types"
)

type UpdatedAtExistsLinter struct {
	script Script
	err    error
	text   string
}

func NewUpdatedAtExistsLinter(script Script) Rule {
	return &UpdatedAtExistsLinter{script: script}
}

func (l *UpdatedAtExistsLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// 查找是否存在名为 updated_at 的字段
	for _, col := range createStmt.Cols {
		if col.Name != nil && strings.ToLower(col.Name.String()) == UpdatedAt {
			return in, true
		}
	}

	l.err = NewLintError(l.script, l.text, "缺少必要字段: updated_at", func(_ []byte) bool {
		return false
	})

	return in, true
}

func (l *UpdatedAtExistsLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *UpdatedAtExistsLinter) Error() error {
	return l.err
}

type UpdatedAtTypeLinter struct {
	script Script
	err    error
	text   string
}

func NewUpdatedAtTypeLinter(script Script) Rule {
	return &UpdatedAtTypeLinter{script: script}
}

func (l *UpdatedAtTypeLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, col := range createStmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != UpdatedAt {
			continue
		}

		// 检查字段类型
		if strings.ToLower(col.Tp.String()) == types.DateTimeStr {
			return in, true
		}

		l.err = NewLintError(l.script, l.text, "updated_at 类型错误: 应当为 datetime", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte(UpdatedAt))
		})

		return in, true
	}

	return in, true
}

func (l *UpdatedAtTypeLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *UpdatedAtTypeLinter) Error() error {
	return l.err
}

type UpdatedAtDefaultValueLinter struct {
	script Script
	err    error
	text   string
}

func NewUpdatedAtDefaultValueLinter(script Script) Rule {
	return &UpdatedAtDefaultValueLinter{script: script}
}

func (l *UpdatedAtDefaultValueLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// 查找是否存在名为 updated_at 的字段
	for _, col := range createStmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != UpdatedAt {
			continue
		}

		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionDefaultValue {
				if expr, ok := opt.Expr.(*ast.FuncCallExpr); ok && strings.ToUpper(expr.FnName.String()) == "CURRENT_TIMESTAMP" {
					return in, true
				}
			}
		}

		l.err = NewLintError(l.script, l.text, "updated_at DEFAULT VALUE 错误: 应当为 CURRENT_TIMESTAMP", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte("updated_at"))
		})
		return in, true

	}

	return in, true
}

func (l *UpdatedAtDefaultValueLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *UpdatedAtDefaultValueLinter) Error() error {
	return l.err
}

type UpdatedAtOnUpdateLinter struct {
	script Script
	err    error
	text   string
}

func NewUpdatedAtOnUpdateLinter(script Script) Rule {
	return &UpdatedAtOnUpdateLinter{script: script}
}

func (l *UpdatedAtOnUpdateLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// 查找是否存在名为 updated_at 的字段
	for _, col := range createStmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != UpdatedAt {
			continue
		}

		// 检查是否自动跟踪修改时间
		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionOnUpdate {
				if expr, ok := opt.Expr.(*ast.FuncCallExpr); ok && strings.ToUpper(expr.FnName.String()) == "CURRENT_TIMESTAMP" {
					return in, true
				}
			}
		}

		l.err = NewLintError(l.script, l.text, "updated_at 缺少 ON UPDATE option 或 ON UPDATE 值错误: 应当为 ON UPDATE CURRENT_TIMESTAMP",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("updated_at"))
			})

		return in, true
	}

	return in, true
}

func (l *UpdatedAtOnUpdateLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *UpdatedAtOnUpdateLinter) Error() error {
	return l.err
}
