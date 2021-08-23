// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linters

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

type ColumnNameLinter struct {
	baseLinter
}

func NewColumnNameLinter(script script.Script) rules.Rule {
	return &ColumnNameLinter{newBaseLinter(script)}
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
		l.err = linterror.New(l.s, l.text, l.err.(linterror.LintError).Lint,
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
			})
	}()

	if compile := regexp.MustCompile(`^[0-9a-z_]+$`); !compile.Match([]byte(name)) {
		l.err = linterror.LintError{
			ScriptName: l.s.Name(),
			Stmt:       l.text,
			Lint:       "invalid field name: it can only contain lowercase English letters, numbers, and underscores",
			Line:       "",
			LintNo:     0,
		}
		return in, true
	}

	if w := name[0]; '0' <= w && w <= '9' {
		l.err = linterror.LintError{
			ScriptName: l.s.Name(),
			Stmt:       l.text,
			Lint:       "invalid field name: can not start with a number",
			Line:       "",
			LintNo:     0,
		}
		return in, true
	}

	words := strings.Split(name, "_")
	for _, w := range words {
		if _, err := strconv.ParseInt(w, 10, 64); w == "" || err == nil {
			l.err = linterror.LintError{
				ScriptName: l.s.Name(),
				Stmt:       l.text,
				Lint:       "invalid field name: there at least is one English letter between two underscores",
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
