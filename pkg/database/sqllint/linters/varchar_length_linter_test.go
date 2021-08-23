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

const varcharLengthLinterSQL = `
create table some_table (
	name varchar(5001)
);
`

func TestNewVarcharLengthLinter(t *testing.T) {
	linter := sqllint.New(linters.NewVarcharLengthLinter)
	if err := linter.Input([]byte(varcharLengthLinterSQL), "varcharLengthLinterSQL"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	if len(errors) == 0 {
		t.Fatal("fails")
	}
}
