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

package pattern_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"

	"github.com/erda-project/erda/pkg/database/pyorm/pattern"
)

func TestGenMigration(t *testing.T) {
	stmt, err := parser.New().ParseOneStmt(createStmt, "", "")
	if err != nil {
		t.Fatalf("failed to ParseOneStmt, err: %v", err)
	}
	create := stmt.(*ast.CreateTableStmt)
	model, err := pattern.CreateTableStmtToModel(create)
	if err != nil {
		t.Fatalf("failed to CreateTableStmtToModel: %v", err)
	}
	var buf = bytes.NewBuffer(nil)
	if err = pattern.GenModel(buf, *model); err != nil {
		t.Fatalf("failed to GenModel: %v", err)
	}

	var migration = pattern.DeveloperScript{
		Models: []string{buf.String(), buf.String()},
	}
	if err = pattern.GenDeveloperScript(os.Stdout, migration); err != nil {
		t.Fatalf("failed to GenDeveloperScript: %v", err)
	}
}
