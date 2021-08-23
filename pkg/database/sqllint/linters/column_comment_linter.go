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

type ColumnCommentLinter struct {
	baseLinter
}

func NewColumnCommentLinter(script script.Script) rules.Rule {
	return &ColumnCommentLinter{newBaseLinter(script)}
}

func (l *ColumnCommentLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	// AlterTableAlterColumn is always valid for this linter, return
	if spec, ok := in.(*ast.AlterTableSpec); ok && spec.Tp == ast.AlterTableAlterColumn {
		return in, true
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	for _, opt := range col.Options {
		if opt.Tp == ast.ColumnOptionComment {
			return in, true
		}
	}
	l.err = linterror.New(l.s, l.text, "missing necessary column definition option: COMMENT",
		func(line []byte) bool {
			if col.Name == nil {
				return true
			}
			return bytes.Contains(line, bytes.ToLower([]byte(col.Name.String())))
		})
	return in, true
}

func (l *ColumnCommentLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *ColumnCommentLinter) Error() error {
	return l.err
}
