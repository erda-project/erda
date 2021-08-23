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
