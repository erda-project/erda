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
	"github.com/erda-project/erda/pkg/database/sqllint/configuration"
	"github.com/erda-project/erda/pkg/database/sqllint/linters"
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
