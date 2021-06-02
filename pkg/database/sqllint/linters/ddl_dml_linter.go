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

type DDLDMLLinter struct {
	baseLinter
}

func NewDDLDMLLinter(script script.Script) rules.Rule {
	return &DDLDMLLinter{newBaseLinter(script)}
}

func (l *DDLDMLLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case ast.DDLNode, ast.DMLNode:
	default:
		l.err = linterror.New(l.s, l.text,
			"language type error: only support DDL, DML, not support DCL, TCL",
			func(line []byte) bool {
				return true
			})
	}

	return in, true
}

func (l *DDLDMLLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, false
}

func (l *DDLDMLLinter) Error() error {
	return l.err
}
