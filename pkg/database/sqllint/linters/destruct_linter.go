// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
