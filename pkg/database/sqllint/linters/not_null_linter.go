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

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

type NotNullLinter struct {
	baseLinter
}

func NewNotNullLinter(script script.Script) rules.Rule {
	return &NotNullLinter{newBaseLinter(script)}
}

func (l *NotNullLinter) Enter(in ast.Node) (out ast.Node, skip bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	// if not CreateTableStmt, always valid, return and skip
	switch in.(type) {
	case *ast.CreateTableStmt:
		return in, false
	case *ast.ColumnDef:
		out = in
		skip = true
	default:
		return in, true
	}

	col, _ := in.(*ast.ColumnDef)
	for _, opt := range col.Options {
		switch opt.Tp {
		case ast.ColumnOptionNotNull, ast.ColumnOptionPrimaryKey:
			return
		}
	}
	l.err = linterror.New(l.s, l.text, "missing necessary column definition option: NOT NULL",
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
