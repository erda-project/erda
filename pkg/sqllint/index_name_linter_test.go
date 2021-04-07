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

const indexNameLinterSQL = `
create table some_table (
	id bigint,
	primary key id (id)
);

create table some_table (
	first_name varchar(10),
	unique index (first_name)
);

create table some_table (
	last_name varchar(10),
	index (last_name)
);

create table some_table (
	id bigint,
	first_name varchar(10),
	last_name varchar(10),
	primary key pk_id (id),
	unique index uk_first_name (first_name),
	index idx_last_name (last_name)
);
`

func TestNewIndexNameLinter(t *testing.T) {
	linter := sqllint.New(sqllint.NewIndexNameLinter)
	if err := linter.Input([]byte(indexNameLinterSQL), "indexNameLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
