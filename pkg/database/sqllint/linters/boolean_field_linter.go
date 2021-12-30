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

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type booleanFieldLinter struct {
	baseLinter
}

// BooleanFieldLinter 引用规约:
// 1、表达是否概念的字段，必须使用boolean、bool、tinyint(1) 、bit 类型。
// 命名方式应当为系表结构如：is_deleted, is_public，符合表达习惯。
func (hub) BooleanFieldLinter(script script.Script, config sqllint.Config) (sqllint.Rule, error) {
	return &booleanFieldLinter{baseLinter: newBaseLinter(script)}, nil
}

func (l *booleanFieldLinter) Enter(in ast.Node) (ast.Node, bool) {
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
			l.err = linterror.New(l.s, l.text, "布尔类型字段名应当为系表结构, 如 is_deleted, has_child",
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
			l.err = linterror.New(l.s, l.text, "表达是否概念的字段类型应当为 tinyint(1) 或 boolean",
				func(line []byte) bool {
					return bytes.Contains(line, []byte(colName))
				})
			return in, true
		}
	}

	return in, true
}

func (l *booleanFieldLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *booleanFieldLinter) Error() error {
	return l.err
}
