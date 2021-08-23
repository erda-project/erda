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

var keywords = map[string]bool{
	"ALL":   true,
	"ALTER": true,
	"AND":   true,
	"ANY":   true,
	"AS":    true,

	"ENABLE":  true,
	"DISABLE": true,

	"ASC":     true,
	"BETWEEN": true,
	"BY":      true,
	"CASE":    true,
	"CAST":    true,

	"CHECK":      true,
	"CONSTRAINT": true,
	"CREATE":     true,
	"DATABASE":   true,
	"DEFAULT":    true,
	"COLUMN":     true,
	"TABLESPACE": true,
	"PROCEDURE":  true,
	"FUNCTION":   true,

	"DELETE":   true,
	"DESC":     true,
	"DISTINCT": true,
	"DROP":     true,
	"ELSE":     true,
	"EXPLAIN":  true,
	"EXCEPT":   true,

	"END":     true,
	"ESCAPE":  true,
	"EXISTS":  true,
	"FOR":     true,
	"FOREIGN": true,

	"FROM":   true,
	"FULL":   true,
	"GROUP":  true,
	"HAVING": true,
	"IN":     true,

	"INDEX":     true,
	"INNER":     true,
	"INSERT":    true,
	"INTERSECT": true,
	"INTERVAL":  true,

	"INTO": true,
	"IS":   true,
	"JOIN": true,
	"KEY":  true,
	"LEFT": true,

	"LIKE":  true,
	"LOCK":  true,
	"MINUS": true,
	"NOT":   true,

	"NULL":  true,
	"ON":    true,
	"OR":    true,
	"ORDER": true,
	"OUTER": true,

	"PRIMARY":    true,
	"REFERENCES": true,
	"RIGHT":      true,
	"SCHEMA":     true,
	"SELECT":     true,

	"SET":      true,
	"SOME":     true,
	"TABLE":    true,
	"THEN":     true,
	"TRUNCATE": true,

	"UNION":    true,
	"UNIQUE":   true,
	"UPDATE":   true,
	"VALUES":   true,
	"VIEW":     true,
	"SEQUENCE": true,
	"TRIGGER":  true,
	"USER":     true,

	"WHEN":  true,
	"WHERE": true,
	"XOR":   true,

	"OVER": true,
	"TO":   true,
	"USE":  true,

	"REPLACE": true,

	"COMMENT": true,
	"COMPUTE": true,
	"WITH":    true,
	"GRANT":   true,
	"REVOKE":  true,

	// mysql procedure
	"WHILE":   true,
	"DO":      true,
	"DECLARE": true,
	"LOOP":    true,
	"LEAVE":   true,
	"ITERATE": true,
	"REPEAT":  true,
	"UNTIL":   true,
	"OPEN":    true,
	"CLOSE":   true,
	"CURSOR":  true,
	"FETCH":   true,
	"OUT":     true,
	"INOUT":   true,

	"LIMIT": true,

	"DUAL":  true,
	"FALSE": true,
	"IF":    true,
	"KILL":  true,

	"TRUE":     true,
	"BINARY":   true,
	"SHOW":     true,
	"CACHE":    true,
	"ANALYZE":  true,
	"OPTIMIZE": true,
	"ROW":      true,
	"BEGIN":    true,
	"DIV":      true,
	"MERGE":    true,

	// for oceanbase & mysql 5.7
	"PARTITION": true,

	"CONTINUE":  true,
	"UNDO":      true,
	"SQLSTATE":  true,
	"CONDITION": true,
	"MOD":       true,
	"CONTAINS":  true,
	"RLIKE":     true,
	"FULLTEXT":  true,
}

type KeywordsLinter struct {
	baseLinter
}

func NewKeywordsLinter(script script.Script) rules.Rule {
	return &KeywordsLinter{newBaseLinter(script)}
}

func (l *KeywordsLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	switch in.(type) {
	case *ast.CreateTableStmt:
		stmt := in.(*ast.CreateTableStmt)
		name := ddlconv.ExtractCreateName(stmt)
		if name == "" {
			return in, false
		}
		if _, ok := keywords[strings.ToUpper(name)]; ok {
			l.err = linterror.New(l.s, l.text, "invalid table name: can not use MySQL keywords to be table name",
				func(_ []byte) bool {
					return false
				})
			return in, true
		}
	case *ast.ColumnDef:
		col := in.(*ast.ColumnDef)
		name := ddlconv.ExtractColName(col)
		if name == "" {
			return in, false
		}
		if _, ok := keywords[strings.ToUpper(name)]; ok {
			l.err = linterror.New(l.s, l.text, "invalid column name: can not use MySQL keywords to be column name",
				func(line []byte) bool {
					return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(name)))
				})
			return in, true
		}
	default:
		return in, false
	}

	return in, false
}

func (l *KeywordsLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *KeywordsLinter) Error() error {
	return l.err
}
