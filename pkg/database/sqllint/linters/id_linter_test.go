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
