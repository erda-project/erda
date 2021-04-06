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
	"regexp"
	"strconv"
	"strings"

	"github.com/pingcap/parser/ast"
)

type ColumnNameLinter struct {
	script Script
	err    error
	text   string
}

func NewColumnNameLinter(script Script) Rule {
	return &ColumnNameLinter{script: script}
}

func (l *ColumnNameLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	if col.Name == nil {
		return in, true
	}

	name := col.Name.OrigColName()
	if name == "" {
		return in, true
	}

	defer func() {
		if l.err == nil {
			return
		}
		l.err = NewLintError(l.script, l.text, l.err.(LintError).Lint,
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
			})
	}()

	if compile := regexp.MustCompile(`^[0-9a-z_]+$`); !compile.Match([]byte(name)) {
		l.err = LintError{
			ScriptName: l.script.Name(),
			Stmt:       l.text,
			Lint:       "字段名不合法: 只能包含小写英文字母、数字、下划线",
			Line:       "",
			LintNo:     0,
		}
		return in, true
	}

	if w := name[0]; '0' <= w && w <= '9' {
		l.err = LintError{
			ScriptName: l.script.Name(),
			Stmt:       l.text,
			Lint:       "字段名不合法: 不能以数字开头",
			Line:       "",
			LintNo:     0,
		}
		return in, true
	}

	words := strings.Split(name, "_")
	for _, w := range words {
		if _, err := strconv.ParseInt(w, 10, 64); w == "" || err == nil {
			l.err = LintError{
				ScriptName: l.script.Name(),
				Stmt:       l.text,
				Lint:       "字段名不合法: 下划线分割的单词中至少包含一个英文字母",
				Line:       "",
				LintNo:     0,
			}
			return in, true
		}
	}

	return in, true
}

func (l *ColumnNameLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *ColumnNameLinter) Error() error {
	return l.err
}
