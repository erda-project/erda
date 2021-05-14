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
	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/sqllint/linterror"
	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/pkg/sqllint/script"
)

type TableCommentLinter struct {
	baseLinter
}

func NewTableCommentLinter(script script.Script) rules.Rule {
	return &TableCommentLinter{newBaseLinter(script)}
}

func (l *TableCommentLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, opt := range stmt.Options {
		if opt.Tp == ast.TableOptionComment {
			return in, true
		}
	}
	l.err = linterror.New(l.s, l.text, "missing necessary table definition option: COMMENT",
		func(line []byte) bool {
			return false
		})
	return in, true
}

func (l *TableCommentLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *TableCommentLinter) Error() error {
	return l.err
}
