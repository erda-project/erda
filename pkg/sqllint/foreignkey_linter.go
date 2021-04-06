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

package sqllint

import (
	"bytes"

	"github.com/pingcap/parser/ast"
)

type ForeignKeyLinter struct {
	script Script
	err    error
	text   string
}

func NewForeignKeyLinter(script Script) Rule {
	return &ForeignKeyLinter{script: script}
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
		l.err = NewLintError(l.script, l.text, "使用了外键: 一切外键概念必须在应用层表达",
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
