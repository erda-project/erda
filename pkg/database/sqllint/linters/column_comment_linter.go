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

type ColumnCommentLinter struct {
	baseLinter
}

func NewColumnCommentLinter(script script.Script) rules.Rule {
	return &ColumnCommentLinter{newBaseLinter(script)}
}

func (l *ColumnCommentLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	// AlterTableAlterColumn is always valid for this linter, return
	if spec, ok := in.(*ast.AlterTableSpec); ok && spec.Tp == ast.AlterTableAlterColumn {
		return in, true
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	for _, opt := range col.Options {
		if opt.Tp == ast.ColumnOptionComment {
			return in, true
		}
	}
	l.err = linterror.New(l.s, l.text, "missing necessary column definition option: COMMENT",
		func(line []byte) bool {
			if col.Name == nil {
				return true
			}
			return bytes.Contains(line, bytes.ToLower([]byte(col.Name.String())))
		})
	return in, true
}

func (l *ColumnCommentLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *ColumnCommentLinter) Error() error {
	return l.err
}
