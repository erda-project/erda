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

	"github.com/erda-project/erda/pkg/sqllint/linterror"
	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/pkg/sqllint/script"
)

type ForeignKeyLinter struct {
	baseLinter
}

func NewForeignKeyLinter(script script.Script) rules.Rule {
	return &ForeignKeyLinter{newBaseLinter(script)}
}

func (l *ForeignKeyLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	constraint, ok := in.(*ast.Constraint)
	if !ok {
		return in, false
	}

	if constraint.Tp == ast.ConstraintForeignKey {
		l.err = linterror.New(l.s, l.text, "used foreign key: all foreign key concept must be expressed at the application layer not at model layer",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("foreign"))
			})
	}

	return in, false
}

func (l *ForeignKeyLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *ForeignKeyLinter) Error() error {
	return l.err
}
