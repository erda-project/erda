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

const booleanFieldLinterSQL = `
create table some_table (
 	public boolean
);

create table some_table (
	public tinyint(1)
);

create table some_table (
	is_public int
);
`

func TestNewBooleanFieldLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewBooleanFieldLinter)
	if err := linter.Input([]byte(booleanFieldLinterSQL), "booleanFieldLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) != 1 {
		t.Error("failed")
	}
}
