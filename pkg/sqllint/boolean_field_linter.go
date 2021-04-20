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
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type BooleanFieldLinter struct {
	script Script
	err    error
	text   string
}

func NewBooleanFieldLinter(script Script) Rule {
	return &BooleanFieldLinter{script: script}
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
			l.err = NewLintError(l.script, l.text, "表达是否概念的字段, 应当以系动词开头, 如 is_deleted, has_child",
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
			l.err = NewLintError(l.script, l.text, "表达是否概念的字段, 应当为 tinyint(1) 或 bit 类型",
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
