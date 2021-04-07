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
	"strings"

	"github.com/pingcap/parser/ast"
)

type VarcharLengthLinter struct {
	script Script
	err    error
	text   string
}

func NewVarcharLengthLinter(script Script) Rule {
	return &VarcharLengthLinter{script: script}
}

func (l *VarcharLengthLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	if col.Tp != nil &&
		strings.Contains(strings.ToLower(col.Tp.String()), "varchar") &&
		col.Tp.Flen > 5000 {
		l.err = NewLintError(l.script, l.text, "字段类型错误: varchar 类型长度不可 > 5000",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte(col.Tp.String()))

			})
	}

	return in, true
}

func (l *VarcharLengthLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *VarcharLengthLinter) Error() error {
	return l.err
}
