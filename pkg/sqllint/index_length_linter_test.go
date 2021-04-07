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

const indexLengthLinterSQL = `
create table some_table (
	name varchar(101),
	index idx_name (name(200))
);

create table some_table (
	some_text varchar(200),
	index idx_some_text (some_text(100))
);

create table some_table (
	some_text varchar(200),
	index idx_some_text (some_text)
);

create table some_table (
	name varchar(300),
	some_text varchar(500),
	index idx_name (name, some_text)
);
`

func TestNewIndexLengthLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewIndexLengthLinter)
	if err := linter.Input([]byte(indexLengthLinterSQL), "indexLengthLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
