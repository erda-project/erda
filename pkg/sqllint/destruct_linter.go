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

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type DestructLinter struct {
	script Script
	err    error
	text   string
}

func NewDestructLinter(script Script) Rule {
	return &DestructLinter{script: script}
}

func (l *DestructLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case *ast.DropTableStmt, *ast.DropDatabaseStmt, *ast.DropUserStmt,
		*ast.TruncateTableStmt, *ast.RenameTableStmt:
		l.err = NewLintError(l.script, l.text, "破坏性语句: 不可删库、删表、删用户、清表、重命名表",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("drop")) ||
					bytes.Contains(bytes.ToLower(line), []byte("trun")) ||
					bytes.Contains(bytes.ToLower(line), []byte("rename"))
			})
		return in, true
	case *ast.AlterTableSpec:
		spec := in.(*ast.AlterTableSpec)
		if spec.Tp == ast.AlterTableRenameColumn {
			l.err = NewLintError(l.script, l.text, "破坏兼容性的 alter spec: 不可修改字段名",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), []byte("rename"))
				})
			return in, true
		}
		if spec.Tp == ast.AlterTableChangeColumn {
			newName := ddlconv.ExtractAlterTableChangeColNewName(spec)
			oldName := ddlconv.ExtractAlterTableChangeColOldName(spec)
			if newName != "" && oldName != "" && newName != oldName {
				l.err = NewLintError(l.script, l.text, "破坏兼容性的 alter spec: 不可修改字段名",
					func(line []byte) bool {
						return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(newName)))
					})
				return in, true
			}
		}
		if spec.Tp == ast.AlterTableDropColumn {
			l.err = NewLintError(l.script, l.text, "破坏性 alter spec: 不可删除字段",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), []byte("drop"))
				})
			return in, true
		}
		return in, true
	}

	return in, false
}

func (l *DestructLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *DestructLinter) Error() error {
	return l.err
}
