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

package sqllint

import (
	"bytes"
	"strings"

	"github.com/pingcap/parser/ast"
)

type IDExistsLinter struct {
	script Script
	err    error
	stmt   string
}

func NewIDExistsLinter(script Script) Rule {
	return &IDExistsLinter{script: script}
}

func (l *IDExistsLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.stmt == "" || in.Text() != "" {
		l.stmt = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// 检查是否存在名为 id 的键
	for _, col := range stmt.Cols {
		if col.Name != nil && strings.ToLower(col.Name.String()) == ID {
			return in, true
		}
	}

	// 如果没有名为 id 的键, 则 lint
	l.err = NewLintError(l.script, l.stmt, "缺少必要字段: id", func(line []byte) bool {
		return false
	})

	return in, true
}

func (l *IDExistsLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IDExistsLinter) Error() error {
	return l.err
}

type IDTypeLinter struct {
	script Script
	err    error
	stmt   string
}

func NewIDTypeLinter(script Script) Rule {
	return &IDTypeLinter{script: script}
}

func (l *IDTypeLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.stmt == "" || in.Text() != "" {
		l.stmt = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, col := range stmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != ID {
			continue
		}

		// 存在名为 id 的键, 检查类型是否正确
		if strings.Contains(strings.ToLower(col.Tp.String()), "bigint") || strings.Contains(strings.ToLower(col.Tp.String()), "char") {
			return in, true
		}

		l.err = NewLintError(l.script, l.stmt, "id 类型错误: 应当是 bigint 类型或 (var)char 类型", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte("id"))
		})

		return in, true
	}

	return in, true
}

func (l *IDTypeLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IDTypeLinter) Error() error {
	return l.err
}

type IDIsPrimaryLinter struct {
	script Script
	err    error
	stmt   string
}

func NewIDIsPrimaryLinter(script Script) Rule {
	return &IDIsPrimaryLinter{script: script}
}

func (l *IDIsPrimaryLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.stmt == "" || in.Text() != "" {
		l.stmt = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// 检查表约束中是否设置了 id 为主键
	for _, constraint := range stmt.Constraints {
		if constraint.Tp == ast.ConstraintPrimaryKey {
			for _, key := range constraint.Keys {
				if key.Column != nil && key.Column.Name.String() == ID {
					return in, true
				}
			}
		}
	}

	for _, col := range stmt.Cols {
		if col.Name == nil || strings.ToLower(col.Name.String()) != ID {
			continue
		}

		// 检查是否在行定义中设置 id 为主键
		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionPrimaryKey {
				return in, true
			}
		}

		// 表约束和字段约束中都没有设置 id 为主键
		l.err = NewLintError(l.script, l.stmt, "主键错误: id 应当为主键 PRIMARY KEY (id)", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte("id"))
		})
	}

	return in, true
}

func (l *IDIsPrimaryLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IDIsPrimaryLinter) Error() error {
	return l.err
}
