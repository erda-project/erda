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

	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type TableNameLinter struct {
	baseLinter
}

func NewTableNameLinter(script script.Script) rules.Rule {
	return &TableNameLinter{newBaseLinter(script)}
}

func (l *TableNameLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true // 只有 create stmt 才验证表名
	}

	name := ddlconv.ExtractCreateName(stmt)
	if name == "" {
		return in, true
	}

	if compile := regexp.MustCompile(`^[0-9a-z_]{1,}$`); !compile.Match([]byte(name)) {
		l.err = linterror.New(l.s, l.text, "invalid table name: it cant only contain lowercase English letters, numbers, underscores",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
			})
		return in, true
	}

	if w := name[0]; '0' <= w && w <= '9' {
		l.err = linterror.New(l.s, l.text, "invalid table name: cat not start with a number",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
			})
		return in, true
	}

	words := strings.Split(name, "_")
	for _, w := range words {
		if _, err := strconv.ParseInt(w, 10, 64); w == "" || err == nil {
			l.err = linterror.New(l.s, l.text, "invalid table name: there at least is one English letter between two underscores",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
				})
			return in, true
		}
	}

	return in, true
}

func (l *TableNameLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *TableNameLinter) Error() error {
	return l.err
}
