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

package linters_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linters"
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
