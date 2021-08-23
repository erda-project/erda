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
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

const UTF8MB4 = "utf8mb4"

type CharsetLinter struct {
	baseLinter
}

func NewCharsetLinter(script script.Script) rules.Rule {
	return &CharsetLinter{newBaseLinter(script)}
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
		if opt.Tp == ast.TableOptionCharset && strings.EqualFold(opt.StrValue, UTF8MB4) {
			return in, true
		}
	}
	l.err = linterror.New(l.s, l.text, "table charset error, it should be CHARSET = utf8mb4",
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
