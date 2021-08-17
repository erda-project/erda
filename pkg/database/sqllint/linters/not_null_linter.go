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
)

type NotNullLinter struct {
	baseLinter
}

func NewNotNullLinter(script script.Script) rules.Rule {
	return &NotNullLinter{newBaseLinter(script)}
}

func (l *NotNullLinter) Enter(in ast.Node) (out ast.Node, skip bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	// if not CreateTableStmt, always valid, return and skip
	switch in.(type) {
	case *ast.CreateTableStmt:
		return in, false
	case *ast.ColumnDef:
		out = in
		skip = true
	default:
		return in, true
	}

	col, _ := in.(*ast.ColumnDef)
	for _, opt := range col.Options {
		switch opt.Tp {
		case ast.ColumnOptionNotNull, ast.ColumnOptionPrimaryKey:
			return
		}
	}
	l.err = linterror.New(l.s, l.text, "missing necessary column definition option: NOT NULL",
		func(line []byte) bool {
			if col.Name == nil {
				return false
			}
			return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(col.Name.String())))
		})
	return
}

func (l *NotNullLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *NotNullLinter) Error() error {
	return l.err
}
