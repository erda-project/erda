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
	"github.com/erda-project/erda/pkg/sqllint/configuration"
	"github.com/erda-project/erda/pkg/sqllint/linters"
)

const columnCommentLinterSQL = `
create table some_table (
	name varchar(101)
);

alter table some_table
	add name varchar(101)
;
`

const alterTableAlterColumnSetDefault = `
alter table t1 alter column col_a set default 0;
`

func TestNewCommentLinter(t *testing.T) {
	linter := sqllint.New(linters.NewColumnCommentLinter)
	if err := linter.Input([]byte(columnCommentLinterSQL), "columnCommentLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors["columnCommentLinterSQL [lints]"]) != 2 {
		t.Fatal("failed")
	}
}

func TestNewCommentLinter2(t *testing.T) {
	linter := sqllint.New(configuration.DefaultRulers()...)
	if err := linter.Input([]byte(alterTableAlterColumnSetDefault), alterTableAlterColumnSetDefault); err != nil {
		t.Fatal(err)
	}

	if len(linter.Errors()) > 0 {
		t.Log(linter.Report())
		t.Log(linter.Errors())
		t.Fatal()
	}
}
