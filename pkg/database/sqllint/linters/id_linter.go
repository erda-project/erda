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

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

const (
	id = "id"
)

type IDExistsLinter struct {
	baseLinter
}

func NewIDExistsLinter(script script.Script) rules.Rule {
	return &IDExistsLinter{newBaseLinter(script)}
}

func (l *IDExistsLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// check id
	for _, col := range stmt.Cols {
		if col.Name != nil && strings.EqualFold(col.Name.String(), id) {
			return in, true
		}
	}

	l.err = linterror.New(l.s, l.text, "missing necessary field: id", func(line []byte) bool {
		return false
	})

	return in, true
}

func (l *IDExistsLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IDExistsLinter) Error() error {
	return l.err
}

type IDTypeLinter struct {
	baseLinter
}

func NewIDTypeLinter(script script.Script) rules.Rule {
	return &IDTypeLinter{newBaseLinter(script)}
}

func (l *IDTypeLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, col := range stmt.Cols {
		if col.Name == nil || !strings.EqualFold(col.Name.String(), id) {
			continue
		}

		// check id column type
		if strings.Contains(strings.ToLower(col.Tp.String()), "bigint") ||
			strings.Contains(strings.ToLower(col.Tp.String()), "char") {
			return in, true
		}

		l.err = linterror.New(l.s, l.text, "type error: id type should be BIGINT or (VAR)CHAR", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte("id"))
		})

		return in, true
	}

	return in, true
}

func (l *IDTypeLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IDTypeLinter) Error() error {
	return l.err
}

type IDIsPrimaryLinter struct {
	baseLinter
}

func NewIDIsPrimaryLinter(script script.Script) rules.Rule {
	return &IDIsPrimaryLinter{newBaseLinter(script)}
}

func (l *IDIsPrimaryLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// check id column is whether primary key or not
	for _, constraint := range stmt.Constraints {
		if constraint.Tp == ast.ConstraintPrimaryKey {
			for _, key := range constraint.Keys {
				if key.Column != nil && key.Column.Name.String() == id {
					return in, true
				}
			}
		}
	}

	for _, col := range stmt.Cols {
		if col.Name == nil || !strings.EqualFold(col.Name.String(), id) {
			continue
		}

		// check id column is whether defined to be primary key in ColDef
		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionPrimaryKey {
				return in, true
			}
		}

		// check id column is whether defined bo be primary key in constraint
		l.err = linterror.New(l.s, l.text, "primary key error: it should be PRIMARY KEY (id)", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte("id"))
		})
	}

	return in, true
}

func (l *IDIsPrimaryLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IDIsPrimaryLinter) Error() error {
	return l.err
}
