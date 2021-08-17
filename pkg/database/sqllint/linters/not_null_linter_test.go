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

	"github.com/erda-project/erda/pkg/database/sqllint"
	"github.com/erda-project/erda/pkg/database/sqllint/linters"
)

const notNullLinterTest = `
create table some_table (
	name varchar(101)
);
`

const notNullLinterTest2 = `
create table t (
	name varchar(225) not null
);
`

const notNullLinterTestAlterTable = `
ALTER TABLE dice_api_access
    ADD default_sla_id BIGINT COMMENT 'default SLA id';
`

func TestNewNotNullLinter(t *testing.T) {
	linter := sqllint.New(linters.NewNotNullLinter)
	if err := linter.Input([]byte(notNullLinterTest), "notNullLinterTest"); err != nil {
		t.Fatal(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}

	linter = sqllint.New(linters.NewNotNullLinter)
	if err := linter.Input([]byte(notNullLinterTestAlterTable), "notNullLinterTestAlterTable"); err != nil {
		t.Fatal(err)
	}
	errors = linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) > 0 {
		t.Fatal("failed")
	}

	linter = sqllint.New(linters.NewNotNullLinter)
	if err := linter.Input([]byte(notNullLinterTest2), "notNullLinterTest2"); err != nil {
		t.Fatal(err)
	}
	errors = linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) > 0 {
		t.Fatal("failed")
	}
}
