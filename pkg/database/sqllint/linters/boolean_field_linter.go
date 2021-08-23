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
	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type BooleanFieldLinter struct {
	baseLinter
}

func NewBooleanFieldLinter(script script.Script) rules.Rule {
	return &BooleanFieldLinter{baseLinter: newBaseLinter(script)}
}

func (l *BooleanFieldLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	colName := ddlconv.ExtractColName(col)
	colType := ddlconv.ExtractColType(col)
	switch colType {
	case "bool", "boolean", "tinyint(1)", "bit":
		if !(strings.HasPrefix(colName, "is_") || strings.HasPrefix(colName, "has_")) {
			l.err = linterror.New(l.s, l.text, "boolean field should start with linking-verb, e.g. is_deleted, has_child",
				func(line []byte) bool {
					return bytes.Contains(line, []byte(colName))
				})
			return in, true
		}
	}

	if strings.HasPrefix(colName, "is_") || strings.HasPrefix(colName, "has_") {
		switch colType {
		case "bool", "boolean", "tinyint(1)", "bit":
			return in, true
		default:
			l.err = linterror.New(l.s, l.text, "boolean field type should be tinyint(1) or boolean",
				func(line []byte) bool {
					return bytes.Contains(line, []byte(colName))
				})
			return in, true
		}
	}

	return in, true
}

func (l *BooleanFieldLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *BooleanFieldLinter) Error() error {
	return l.err
}
