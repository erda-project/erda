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

// ExplicitCollationLinter lints if user set the collation explicitly
type ExplicitCollationLinter struct {
	baseLinter
}

// NewExplicitCollationLinter returns an ExplicitCollationLinter
func NewExplicitCollationLinter(script script.Script) rules.Rule {
	return &ExplicitCollationLinter{baseLinter: newBaseLinter(script)}
}

func (l *ExplicitCollationLinter) Enter(in ast.Node) (node ast.Node, skipChildren bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// check if the collation is in the table options
	for _, opt := range stmt.Options {
		if opt == nil {
			continue
		}
		if opt.Tp == ast.TableOptionCollate {
			l.err = linterror.New(l.s, l.text, "should not set the collation in CreateTableStmt", func(line []byte) bool {
				return bytes.Contains(line, []byte(opt.StrValue))
			})
			return in, true
		}
	}

	// check if the collation is in the columns' options
	for _, col := range stmt.Cols {
		if col == nil {
			continue
		}
		if col.Tp != nil && col.Tp.Charset != "" {
			l.err = linterror.New(l.s, l.text, "should not set the character in ColumnDef", func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte("character"))) && bytes.Contains(line, []byte(col.Tp.Charset))
			})
			return in, true
		}
		for _, opt := range col.Options {
			if opt == nil {
				continue
			}
			if opt.Tp == ast.ColumnOptionCollate {
				l.err = linterror.New(l.s, l.text, "should not set the collation in ColumnDef", func(line []byte) bool {
					return bytes.Contains(line, []byte(opt.StrValue))
				})
				return in, true
			}
		}
	}

	return in, true
}

func (l *ExplicitCollationLinter) Leave(in ast.Node) (node ast.Node, ok bool) {
	return in, true
}
