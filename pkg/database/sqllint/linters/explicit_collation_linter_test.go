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

func TestNewExplicitCollationLinter(t *testing.T) {
	var (
		sqlA = `
create table t0 (
	left_key varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'asset id'
);
`
		sqlB = `
create table t0 (
	left_key varchar(1024)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API';
`
		sqlC = `
create table t0 (
	left_key varchar(1024) character set utf8mb4  COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'asset id'
)
`

		sqlD = `
create table t0 (
	left_key varchar(1024)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API';
`
	)

	linter := sqllint.New(linters.NewExplicitCollationLinter)
	if err := linter.Input([]byte(sqlA), "sqlA"); err != nil {
		t.Fatalf("failed to Input sqlA to linter: %v", err)
	}
	if errs := linter.Errors(); len(errs) == 0 {
		t.Fatal("failed to lint, there should be some errors")
	} else {
		t.Logf("sqlA's error: %+v", errs)
	}

	linter = sqllint.New(linters.NewExplicitCollationLinter)
	if err := linter.Input([]byte(sqlB), "sqlB"); err != nil {
		t.Fatalf("failed to Input sqlB to linter: %v", err)
	}
	if errs := linter.Errors(); len(errs) == 0 {
		t.Fatal("failed to lint, there should be some errors")
	} else {
		t.Logf("sqlB's error: %+v", errs)
	}

	linter = sqllint.New(linters.NewExplicitCollationLinter)
	if err := linter.Input([]byte(sqlC), "sqlC"); err != nil {
		t.Fatalf("failed to Input sqlC to linter: %v", err)
	}
	if errs := linter.Errors(); len(errs) == 0 {
		t.Fatal("failed to lint, there should be some errors")
	} else {
		t.Logf("sqlC's error: %v", errs)
	}

	linter = sqllint.New(linters.NewExplicitCollationLinter)
	if err := linter.Input([]byte(sqlD), "sqlD"); err != nil {
		t.Fatalf("failed to Input sqlC to linter: %v", err)
	}
	if errs := linter.Errors(); len(errs) != 0 {
		t.Fatalf("failed to lint, there should be no errors, but errors: %+v", errs)
	}

}
