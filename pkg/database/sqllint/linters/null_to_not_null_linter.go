//  Copyright (c) 2021 Terminus, Inc.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package linters

import (
	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/schema"
	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
)

type NullToNotNullLinter struct {
	baseLinter
	tableName string
	schema    *schema.Schema
}

func (l *NullToNotNullLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch stmt := in.(type) {
	case *ast.AlterTableStmt:
		l.tableName = stmt.Table.Name.String()
		return in, false
	case *ast.AlterTableSpec:
		switch stmt.Tp {
		case ast.AlterTableChangeColumn, ast.AlterTableAlterColumn, ast.AlterTableModifyColumn:
			if len(stmt.NewColumns) == 0 {
				return in, true
			}
			if !hasOption(ast.ColumnOptionNotNull, stmt.NewColumns[0].Options) {
				return in, true
			}
			definition := l.schema.TableDefinitions[l.tableName]
			if definition == nil {
				return in, true
			}
			oldCol := getCol(stmt.OldColumnName.String(), definition.CreateStmt.Cols)
			if oldCol == nil {
				return in, true
			}
			if hasOption(ast.ColumnOptionNull, oldCol.Options) {
				l.err = linterror.New(l.s, l.text, "can not convert 'NULL' column to 'NOT NULL'",
					func(line []byte) bool {
						return true
					})
				return in, true
			}
			return in, true
		}
	default:
		return in, true
	}
	return in, true
}

func (l *NullToNotNullLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *NullToNotNullLinter) Error() error {
	return l.err
}

func (l *NullToNotNullLinter) SetSchema(schema *schema.Schema) {
	l.schema = schema
}

func hasOption(tp ast.ColumnOptionType, options []*ast.ColumnOption) bool {
	for _, opt := range options {
		switch opt.Tp {
		case tp:
			return true
		}
	}
	return false
}

func getCol(name string, cols []*ast.ColumnDef) *ast.ColumnDef {
	for _, col := range cols {
		if name == col.Name.String() {
			return col
		}
	}
	return nil
}
