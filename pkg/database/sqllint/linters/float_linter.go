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
	"strings"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqllint/linterror"
	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqllint/script"
	"github.com/erda-project/erda/pkg/swagger/ddlconv"
)

type FloatDoubleLinter struct {
	baseLinter
}

func NewFloatDoubleLinter(script script.Script) rules.Rule {
	return &FloatDoubleLinter{newBaseLinter(script)}
}

func (l *FloatDoubleLinter) Enter(in ast.Node) (ast.Node, bool) {
	if l.text == "" || in.Text() != "" {
		l.text = in.Text()
	}

	col, ok := in.(*ast.ColumnDef)
	if !ok {
		return in, false
	}

	t := ddlconv.ExtractColType(col)
	t = strings.ToLower(t)
	if strings.Contains(t, "float") || strings.Contains(t, "double") {
		l.err = linterror.New(l.s, l.text, "field type error: decimal type should use DECIMAL instead of FLOAT OR DOUBLE",
			func(line []byte) bool {
				return bytes.Contains(bytes.ToLower(line), []byte("float")) ||
					bytes.Contains(bytes.ToLower(line), []byte("double"))
			})
	}

	return in, false
}

func (l *FloatDoubleLinter) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func (l *FloatDoubleLinter) Error() error {
	return l.err
}
