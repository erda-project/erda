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

const tableCommentLinterTest = `
CREATE TABLE IF NOT EXISTS dice_api_slas(
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT 'primary key'
);

CREATE TABLE IF NOT EXISTS dice_api_slas(
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT 'primary key'
) comment='this is comment';
`

func TestNewTableCommentLinter(t *testing.T) {
	linter := sqllint.New(linters.NewTableCommentLinter)
	if err := linter.Input([]byte(tableCommentLinterTest), "tableCommentLinterTest"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
	if len(errors) == 0 {
		t.Fatal("failed")
	}
}
