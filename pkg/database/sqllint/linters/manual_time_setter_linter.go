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

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
)

// ManualTimeSetterLinter lints if the user manually set the column created_at, updated_at
type ManualTimeSetterLinter struct {
	baseLinter
}

// NewManualTimeSetterLinter returns a ManualTimeSetterLinter
// ManualTimeSetterLinter lints if the user manually set the column created_at, updated_at
func NewManualTimeSetterLinter(script script.Script) rules.Rule {
	return &ManualTimeSetterLinter{newBaseLinter(script)}
}

func (l *ManualTimeSetterLinter) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	getLint := func(line []byte) bool {
		return bytes.Contains(line, []byte(createAt)) || bytes.Contains(line, []byte(updatedAt))
	}

	switch in.(type) {
	case *ast.InsertStmt:
		columns := in.(*ast.InsertStmt).Columns
		for _, col := range columns {
			if colName := col.Name.String(); strings.EqualFold(colName, createAt) || strings.EqualFold(colName, updatedAt) {
				l.err = linterror.New(l.s, l.text, "can not set created_at or updated_at in INSERT statement", getLint)
				return in, true
			}
		}
	case *ast.UpdateStmt:
		columns := in.(*ast.UpdateStmt).List
		for _, col := range columns {
			if colName := col.Column.String(); strings.EqualFold(colName, createAt) || strings.EqualFold(colName, updatedAt) {
				l.err = linterror.New(l.s, l.text, "can not set created_at or updated_at in UPDATE statement", getLint)
				return in, true
			}
		}
	default:
		return in, true
	}

	return in, true
}

func (l *ManualTimeSetterLinter) Leave(in ast.Node) (out ast.Node, ok bool) {
	return in, true
}
