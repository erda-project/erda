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

package linters

import (
	"bytes"
	"strconv"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/sqllint/linterror"
	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/pkg/sqllint/script"
	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type IndexLengthLinter struct {
	baseLinter
}

func NewIndexLengthLinter(script script.Script) rules.Rule {
	return &IndexLengthLinter{newBaseLinter(script)}
}

func (l *IndexLengthLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	stmt, ok := in.(*ast.CreateTableStmt)
	if !ok {
		return in, true
	}
	var colNames = make(map[string]int, 0)
	for _, col := range stmt.Cols {
		colName := ddlconv.ExtractColName(col)
		if colName == "" {
			continue
		}
		if col.Tp == nil {
			continue
		}
		colNames[colName] = col.Tp.Flen
	}
	for _, c := range stmt.Constraints {
		var length, firstLen int
		for _, key := range c.Keys {
			if key.Length > 0 {
				length += key.Length
				firstLen = key.Length
				continue
			}
			if key.Column == nil {
				continue
			}
			if l, ok := colNames[key.Column.Name.String()]; ok {
				length += l
				firstLen = l
				continue
			}
		}
		if len(c.Keys) == 1 && length*4 > 767 {
			l.err = linterror.New(l.s, l.text, "index length error: single column index length can not bigger than 767",
				func(line []byte) bool {
					firstLenS := strconv.FormatInt(int64(firstLen), 10)
					return bytes.Contains(bytes.ToLower(line), bytes.ToLower([]byte(firstLenS)))
				})
			return in, true
		}
		if len(c.Keys) > 1 && length*4 > 3072 {
			_, num := linterror.CalcLintLine(l.s.Data(), []byte(l.text), func(_ []byte) bool {
				return false
			})
			l.err = linterror.LintError{
				Stmt:   l.text,
				Lint:   "index length error: joint index length can not bigger than 3072",
				LintNo: num,
			}
			return in, true
		}
	}

	return in, true
}

func (l *IndexLengthLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, l.err == nil
}

func (l *IndexLengthLinter) Error() error {
	return l.err
}
