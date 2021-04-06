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
	"strings"

	"github.com/pingcap/parser/ast"
)

const UTF8MB4 = "utf8mb4"

type CharsetLinter struct {
	script Script
	err    error
	text   string
}

func NewCharsetLinter(script Script) Rule {
	return &CharsetLinter{script: script}
}

func (l *CharsetLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, opt := range stmt.Options {
		if opt.Tp == ast.TableOptionCharset && strings.ToLower(opt.StrValue) == UTF8MB4 {
			return in, true
		}
	}
	l.err = NewLintError(l.script, l.text, "表字符集错误: 应当显示声明为 CHARSET = utf8mb4",
		func(line []byte) bool {
			return false
		})
	return in, true
}

func (l *CharsetLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *CharsetLinter) Error() error {
	return l.err
}
