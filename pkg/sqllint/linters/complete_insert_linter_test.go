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

package linters_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
	"github.com/erda-project/erda/pkg/sqllint/linters"
)

func TestNewCompleteInsertLinter(t *testing.T) {
	var (
		insert         = "insert into t1 \nvalues (1, 2, 3)"
		completeInsert = "insert into t2 (col1, col2, col3) \nvalues (1, 2, 3)"
		skip           = "create table t1 (id bigint)"
	)
	linter := sqllint.New(linters.NewCompleteInsertLinter)
	if err := linter.Input([]byte(insert), "insert"); err != nil {
		t.Fatalf("failed to Input data to linter: %v", err)
	}
	if errs := linter.Errors(); len(errs) == 0 {
		t.Fatal("failed to lint, there should be some errors")
	} else {
		t.Logf("%+v", errs)
	}

	linter = sqllint.New(linters.NewCompleteInsertLinter)
	if err := linter.Input([]byte(completeInsert), "completeInsert"); err != nil {
		t.Fatalf("failed to Input data to linter: %v", err)
	}
	if errs := linter.Errors(); len(errs) != 0 {
		t.Logf("%+v", errs)
		t.Fatal("failed to lint, there should be no error")
	}

	linter = sqllint.New(linters.NewCompleteInsertLinter)
	if err := linter.Input([]byte(skip), "skip"); err != nil {
		t.Fatalf("failed to Input data to linter: %v", err)
	}
	if errs := linter.Errors(); len(errs) > 0 {
		t.Log(errs)
		t.Fatalf("failed to lint: %s", "skip")
	}
}
