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

const createTableWithoutID = `
Create Table some_table (
	created_at datetime,
	updated_at datetime
)
`

const createTableWithID = `
create table some_table (
	id bigint
)
`

const createTableWithErrorIDType = `
create table some_table (
	id datetime,
	created_at datetime
)
`

const createTableWithBigintID = `
create table some_table (
	id bigint,
	created_at datetime
)
`

const createTableWithVarcharID = `
create table some_table (
	id varchar(199),
	created_at datetime
)
`

const idIsPrimary = `
create table some_table (
	id bigint primary key,
	uuid bigint
)
`

const idIsNotPrimary = `
create table some_table (
	id bigint,
	uuid bigint primary key
)
`

const primaryIsNotId = `
create table some_table (
	uuid bigint primary key,
	id bigint
)
`

func TestNewIDExistsLinter(t *testing.T) {
	t.Run("testNewIDExistsLinter_WithoutID", testnewidexistslinterWithoutid)
	t.Run("testNewIDExistsLinterWithID", testNewIDExistsLinterWithID)
}

func testnewidexistslinterWithoutid(t *testing.T) {
	linter := sqllint.New(linters.NewIDExistsLinter)
	if err := linter.Input([]byte(createTableWithoutID), "createTableWithouID"); err != nil {
		t.Error(err)
	}

	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("fails")
	}
}

func testNewIDExistsLinterWithID(t *testing.T) {
	linter := sqllint.New(linters.NewIDExistsLinter)
	if err := linter.Input([]byte(createTableWithID), "createTableWithID"); err != nil {
		t.Error(err)
	}

	errors := linter.Errors()
	if len(errors) > 0 {
		t.Fatal("fails")
	}
}

func TestNewIDTypeLinter(t *testing.T) {
	t.Run("testNewIDTypeLinterWithErrorIDType", testNewIDTypeLinterWithErrorIDType)
	t.Run("testNewIDTypeLinterWithBigintID", testNewIDTypeLinterWithBigintID)
	t.Run("testNewIDTypeLinterWithVarcharID", testNewIDTypeLinterWithVarcharID)
}

func testNewIDTypeLinterWithErrorIDType(t *testing.T) {
	linter := sqllint.New(linters.NewIDTypeLinter)
	if err := linter.Input([]byte(createTableWithErrorIDType), "createTableWithErrorIDType"); err != nil {
		t.Error(err)
	}

	if errors := linter.Errors(); len(errors) == 0 {
		t.Fatal("fails")
	}
}

func testNewIDTypeLinterWithBigintID(t *testing.T) {
	linter := sqllint.New(linters.NewIDTypeLinter)
	if err := linter.Input([]byte(createTableWithBigintID), "createTableWithBigintID"); err != nil {
		t.Error(err)
	}
	if errors := linter.Errors(); len(errors) > 0 {
		t.Error("fails")
	}
}

func testNewIDTypeLinterWithVarcharID(t *testing.T) {
	linter := sqllint.New(linters.NewIDTypeLinter)
	if err := linter.Input([]byte(createTableWithVarcharID), "createTableWithVarcharID"); err != nil {
		t.Error(err)
	}
	if errors := linter.Errors(); len(errors) > 0 {
		t.Error("fails")
	}
}

func TestNewIDIsPrimaryLinter(t *testing.T) {
	t.Run("testNewIDIsPrimaryLinterIDIsPrimary", testNewIDIsPrimaryLinterIDIsPrimary)
	t.Run("testNewIDIsPrimaryLinterIDIsNotPrimary", testNewIDIsPrimaryLinterIDIsNotPrimary)
	t.Run("testNewIDIsPrimaryLinterPrimaryIsNotID", testNewIDIsPrimaryLinterPrimaryIsNotID)
}

func testNewIDIsPrimaryLinterIDIsPrimary(t *testing.T) {
	linter := sqllint.New(linters.NewIDIsPrimaryLinter)
	if err := linter.Input([]byte(idIsPrimary), "idIsPrimary"); err != nil {
		t.Fatal(err)
	}
	if errors := linter.Errors(); len(errors) > 0 {
		t.Fatal("fails")
	}
}

func testNewIDIsPrimaryLinterIDIsNotPrimary(t *testing.T) {
	linter := sqllint.New(linters.NewIDIsPrimaryLinter)
	if err := linter.Input([]byte(idIsNotPrimary), "idIsNotPrimary"); err != nil {
		t.Fatal(err)
	}
	if errors := linter.Errors(); len(errors) == 0 {
		t.Fatal("fails")
	}
}

func testNewIDIsPrimaryLinterPrimaryIsNotID(t *testing.T) {
	linter := sqllint.New(linters.NewIDIsPrimaryLinter)
	if err := linter.Input([]byte(primaryIsNotId), "primaryIsNotId"); err != nil {
		t.Fatal(err)
	}
	if errors := linter.Errors(); len(errors) == 0 {
		t.Fatal("fails")
	}
}
