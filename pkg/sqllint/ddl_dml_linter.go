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
	"github.com/pingcap/parser/ast"
)

type DDLDMLLinter struct {
	script Script
	err    error
	text   string
}

func NewDDLDMLLinter(script Script) Rule {
	return &DDLDMLLinter{script: script}
}

func (l *DDLDMLLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case ast.DDLNode, ast.DMLNode:
	default:
		l.err = NewLintError(l.script, l.text,
			"语言类型错误: 只能包含数据定义语言(DDL)、数据操作语言(DML), 不可以包含数据库操作语言(DCL)、事务控制语言(TCL)",
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
