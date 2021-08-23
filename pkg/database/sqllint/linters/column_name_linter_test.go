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

const columnNameLinterSQL = `
create table some_table (
	姓名 varchar(100)
);

create table some_table (
	BigBang varchar(100)
);

create table some_table (
	3d_school varchar(100)
);

create table some_table (
	level_3_vip varchar(100)
);
`

func TestNewColumnNameLinter(t *testing.T) {
	linter := sqllint.New(linters.NewColumnNameLinter)
	if err := linter.Input([]byte(columnNameLinterSQL), "columnNameLinterSQL"); err != nil {
		t.Error(err)
	}

	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors["columnNameLinterSQL [lints]"]) != 4 {
		t.Fatal("failed", len(errors["columnNameLinterSQL"]))
	}
}
