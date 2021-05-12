// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package linters

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/tidb/types"

	"github.com/erda-project/erda/pkg/sqllint/linterror"
	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/pkg/sqllint/script"
)

const (
	updatedAt = "updated_at"
)

type UpdatedAtExistsLinter struct {
	baseLinter
}

func NewUpdatedAtExistsLinter(script script.Script) rules.Rule {
	return &UpdatedAtExistsLinter{newBaseLinter(script)}
}

func (l *UpdatedAtExistsLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// check updated_at
	for _, col := range createStmt.Cols {
		if col.Name != nil && strings.EqualFold(col.Name.String(), updatedAt) {
			return in, true
		}
	}

	l.err = linterror.New(l.s, l.text, "missing necessary column: updated_at", func(_ []byte) bool {
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
	baseLinter
}

func NewUpdatedAtTypeLinter(script script.Script) rules.Rule {
	return &UpdatedAtTypeLinter{newBaseLinter(script)}
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
		if col.Name == nil || !strings.EqualFold(col.Name.String(), updatedAt) {
			continue
		}

		// 检查字段类型
		if strings.EqualFold(col.Tp.String(), types.DateTimeStr) {
			return in, true
		}

		l.err = linterror.New(l.s, l.text, "column type error: updated_at should be DATETIME", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte(updatedAt))
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
	baseLinter
}

func NewUpdatedAtDefaultValueLinter(script script.Script) rules.Rule {
	return &UpdatedAtDefaultValueLinter{newBaseLinter(script)}
}

func (l *UpdatedAtDefaultValueLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// check updated_at
	for _, col := range createStmt.Cols {
		if col.Name == nil || !strings.EqualFold(col.Name.String(), updatedAt) {
			continue
		}

		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionDefaultValue {
				if expr, ok := opt.Expr.(*ast.FuncCallExpr); ok &&
					strings.EqualFold(expr.FnName.String(), "CURRENT_TIMESTAMP") {
					return in, true
				}
			}
		}

		l.err = linterror.New(l.s, l.text, "default value error: updated_at defaults CURRENT_TIMESTAMP", func(line []byte) bool {
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
	baseLinter
}

func NewUpdatedAtOnUpdateLinter(script script.Script) rules.Rule {
	return &UpdatedAtOnUpdateLinter{newBaseLinter(script)}
}

func (l *UpdatedAtOnUpdateLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// check updated_at
	for _, col := range createStmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != updatedAt {
			continue
		}

		// check default
		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionOnUpdate {
				if expr, ok := opt.Expr.(*ast.FuncCallExpr); ok &&
					strings.EqualFold(expr.FnName.String(), "CURRENT_TIMESTAMP") {
					return in, true
				}
			}
		}

		l.err = linterror.New(l.s, l.text, "missing necessary column option: ON UPDATE CURRENT_TIMESTAMP for updated_at",
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
