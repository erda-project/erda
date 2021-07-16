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

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type DestructLinter struct {
	baseLinter
}

func NewDestructLinter(script script.Script) rules.Rule {
	return &DestructLinter{newBaseLinter(script)}
}

func (l *DestructLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case *ast.DropTableStmt, *ast.DropDatabaseStmt, *ast.DropUserStmt,
		*ast.TruncateTableStmt, *ast.RenameTableStmt:
		l.err = linterror.New(l.s, l.text, "destructive operation: can not drop database, drop table, drop user, truncate table, rename table ...",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("drop")) ||
					bytes.Contains(bytes.ToLower(line), []byte("trun")) ||
					bytes.Contains(bytes.ToLower(line), []byte("rename"))
			})
		return in, true
	case *ast.AlterTableSpec:
		spec := in.(*ast.AlterTableSpec)
		if spec.Tp == ast.AlterTableRenameColumn {
			l.err = linterror.New(l.s, l.text, "break compatibility: AlterTableRenameColumn",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), []byte("rename"))
				})
			return in, true
		}
		if spec.Tp == ast.AlterTableChangeColumn {
			newName := ddlconv.ExtractAlterTableChangeColNewName(spec)
			oldName := ddlconv.ExtractAlterTableChangeColOldName(spec)
			if newName != "" && oldName != "" && newName != oldName {
				l.err = linterror.New(l.s, l.text, "break compatibility: AlterTableChangeColumnName",
					func(line []byte) bool {
						return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(newName)))
					})
				return in, true
			}
		}
		if spec.Tp == ast.AlterTableDropColumn {
			l.err = linterror.New(l.s, l.text, "break compatibility: AlterTableDropColumn",
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
