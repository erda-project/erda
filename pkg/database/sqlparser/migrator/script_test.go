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

package migrator_test

import (
	"testing"

	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/sqlparser/migrator"
)

func TestNewScript(t *testing.T) {
	script, err := migrator.NewScript("..", "testdata/dice_base.sql")
	if err != nil {
		t.Fatal(err)
	}

	for _, dml := range script.DMLNodes() {
		stmt, ok := dml.(*ast.SetStmt)
		if ok {
			t.Log(stmt.Text())
		}
	}
}

func TestScript_IsEmpty(t *testing.T) {
	var s = new(migrator.Script)
	if !s.IsEmpty() {
		t.Fatal("the script is empty")
	}

	s.Rawtext = []byte("select 1 from t;")
	if s.IsEmpty() {
		t.Fatal("the script is not empty")
	}
}