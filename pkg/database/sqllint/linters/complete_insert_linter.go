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

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

// CompleteInsertLinter lints if the INSERT statement is complete
type CompleteInsertLinter struct {
	baseLinter
}

// NewCompleteInsertLinter returns a CompleteInsertLinter
// CompleteInsertLinter lint if the INSERT statement is complete-insert.
//
// e.g. NOT OK:
// INSERT INTO table_name
// VALUES (value1,value2,value3,...);
//
// OK:
// INSERT INTO table_name (column1,column2,column3,...)
// VALUES (value1,value2,value3,...);
func NewCompleteInsertLinter(script script.Script) rules.Rule {
	return &CompleteInsertLinter{baseLinter: newBaseLinter(script)}
}

func (l *CompleteInsertLinter) Enter(in ast.Node) (node ast.Node, skipChildren bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.InsertStmt)
	if !ok {
		return in, true
	}

	if len(stmt.Columns) == 0 {
		l.err = linterror.New(l.s, l.text, "INSERT stmt should be complete-insert, you have to specify column names",
			func(line []byte) bool {
				return false
			})
	}

	return in, true
}

func (l *CompleteInsertLinter) Leave(in ast.Node) (node ast.Node, ok bool) {
	return in, true
}
