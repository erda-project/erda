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

	"github.com/erda-project/erda/pkg/sqllint/linterror"
	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/pkg/sqllint/script"
)

// ManualTimeSetterLinter lints if the user manually set the column created_at, updated_at
type ManualTimeSetterLinter struct {
	baseLinter
}

// NewManualTimeSetterLinter returns a ManualTimeSetterLinter
// ManualTimeSetterLinter lints if the user manually set the column created_at, updated_at
func NewManualTimeSetterLinter(script script.Script) rules.Rule {
	return &ManualTimeSetterLinter{newBaseLinter(script)}
}

func (l *ManualTimeSetterLinter) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	getLint := func(line []byte) bool {
		return bytes.Contains(line, []byte(createAt)) || bytes.Contains(line, []byte(updatedAt))
	}

	switch in.(type) {
	case *ast.InsertStmt:
		columns := in.(*ast.InsertStmt).Columns
		for _, col := range columns {
			if colName := col.Name.String(); strings.EqualFold(colName, createAt) || strings.EqualFold(colName, updatedAt) {
				l.err = linterror.New(l.s, l.text, "can not set created_at or updated_at in INSERT statement", getLint)
				return in, true
			}
		}
	case *ast.UpdateStmt:
		columns := in.(*ast.UpdateStmt).List
		for _, col := range columns {
			if colName := col.Column.String(); strings.EqualFold(colName, createAt) || strings.EqualFold(colName, updatedAt) {
				l.err = linterror.New(l.s, l.text, "can not set created_at or updated_at in UPDATE statement", getLint)
				return in, true
			}
		}
	default:
		return in, true
	}

	return in, true
}

func (l *ManualTimeSetterLinter) Leave(in ast.Node) (out ast.Node, ok bool) {
	return in, true
}
