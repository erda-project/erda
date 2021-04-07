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

package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
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
	linter := sqllint.New(sqllint.NewColumnNameLinter)
	if err := linter.Input([]byte(columnNameLinterSQL), "columnNameLinterSQL"); err != nil {
		t.Error(err)
	}

	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors["columnNameLinterSQL [lints]"]) != 4 {
		t.Fatal("failed", len(errors["columnNameLinterSQL"]))
	}
}
