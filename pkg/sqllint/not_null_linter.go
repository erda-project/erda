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

type NotNullLinter struct {
	script Script
	err    error
	text   string
}

func NewNotNullLinter(script Script) Rule {
	return &NotNullLinter{script: script}
}

func (l *NotNullLinter) Enter(in ast.Node) (out ast.Node, skip bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	out = in

	col, skip := in.(*ast.ColumnDef)
	if !skip {
		return
	}

	for _, opt := range col.Options {
		switch opt.Tp {
		case ast.ColumnOptionNotNull, ast.ColumnOptionPrimaryKey:
			return
		}
	}
	l.err = NewLintError(l.script, l.text, "字段定义子句缺少必要的 option: 应当显示声明 NOT NULL",
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
