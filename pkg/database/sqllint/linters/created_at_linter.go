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
	"strings"

	"github.com/pingcap/parser/ast"
	"github.com/pingcap/tidb/types"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

const (
	createAt = "created_at"
)

type CreatedAtExistsLinter struct {
	baseLinter
}

func NewCreatedAtExistsLinter(script script.Script) rules.Rule {
	return &CreatedAtExistsLinter{newBaseLinter(script)}
}

func (l *CreatedAtExistsLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	// look for "created_at"
	for _, col := range createStmt.Cols {
		if col.Name != nil && strings.EqualFold(col.Name.String(), createAt) {
			return in, true
		}
	}

	// if no "created_at"
	l.err = linterror.New(l.s, l.text, "missing necessary field: created_at", func(line []byte) bool {
		return false
	})

	return in, true
}

func (l *CreatedAtExistsLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *CreatedAtExistsLinter) Error() error {
	return l.err
}

type CreatedAtTypeLinter struct {
	baseLinter
}

func NewCreatedAtTypeLinter(script script.Script) rules.Rule {
	return &CreatedAtTypeLinter{newBaseLinter(script)}
}

func (l *CreatedAtTypeLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, col := range createStmt.Cols {
		if col.Name == nil || !strings.EqualFold(col.Name.String(), createAt) {
			continue
		}

		// check type
		if strings.EqualFold(col.Tp.String(), types.DateTimeStr) {
			return in, true
		}

		l.err = linterror.New(l.s, l.text, "type error: created_at should be DATETIME", func(line []byte) bool {
			return bytes.Contains(bytes.ToLower(line), []byte("created_at"))
		})

		return in, true
	}

	return in, true
}

func (l *CreatedAtTypeLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *CreatedAtTypeLinter) Error() error {
	return l.err
}

type CreatedAtDefaultValueLinter struct {
	baseLinter
}

func NewCreatedAtDefaultValueLinter(script script.Script) rules.Rule {
	return &CreatedAtDefaultValueLinter{newBaseLinter(script)}
}

func (l *CreatedAtDefaultValueLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	createStmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}

	for _, col := range createStmt.Cols {
		if col.Name == nil || !strings.EqualFold(col.Name.String(), createAt) {
			continue
		}

		// check default value
		for _, opt := range col.Options {
			if opt.Tp == ast.ColumnOptionDefaultValue {
				if expr, ok := opt.Expr.(*ast.FuncCallExpr); ok &&
					strings.EqualFold(expr.FnName.String(), "CURRENT_TIMESTAMP") {
					return in, true
				}
			}
		}

		l.err = linterror.New(l.s, l.text, "DEFAULT VALUE error: created_at defaults CURRENT_TIMESTAMP",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("created_at"))
			})

		return in, true
	}

	return in, true
}

func (l *CreatedAtDefaultValueLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *CreatedAtDefaultValueLinter) Error() error {
	return l.err
}
