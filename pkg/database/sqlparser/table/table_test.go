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

package table_test

import (
	"testing"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqlparser/table"
)

func TestTable_Append(t *testing.T) {
	var (
		tbl  table.Table
		node = new(ast.CreateTableStmt)
	)
	tbl.Append(node)
	nodes := tbl.Nodes()
	if len(nodes) != 1 {
		t.Fatal("error")
	}
	if _, ok := nodes[0].(*ast.CreateTableStmt); !ok {
		t.Fatal("error")
	}
}
