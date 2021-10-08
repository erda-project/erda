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

func TestNewManualTimeSetterLinter(t *testing.T) {
	var (
		insertCreatedAt = "insert into t1 (created_at) values ('2021-01-01 00:00:00')"
		insertUpdatedAt = "insert into t1 (updated_at) values ('2021-01-01 00:00:00')"
		updateCreatedAt = "update t1 set created_at = '2021-01-01 00:00:00'"
		updateUpdatedAt = "update t2 set updated_at = '2021-01-01 00:00:00'"
		skip            = "create table t1 (id bigint)"
		skip2           = "insert into t1 (col1) values (0)"
		m               = map[string][]byte{
			"insertCreatedAt": []byte(insertCreatedAt),
			"insertUpdatedAt": []byte(insertUpdatedAt),
			"updateCreatedAt": []byte(updateCreatedAt),
			"updateUpdatedAt": []byte(updateUpdatedAt),
		}
		skips = map[string][]byte{
			"skip":  []byte(skip),
			"skip2": []byte(skip2),
		}
	)

	for name, data := range m {
		linter := sqllint.New(linters.NewManualTimeSetterLinter)
		if err := linter.Input(data, name); err != nil {
			t.Fatalf("failed to Input: %s", name)
		}
		if errs := linter.Errors(); len(errs) == 0 {
			t.Fatalf("failed to lint: %s", name)
		}
	}

	for name, data := range skips {
		linter := sqllint.New(linters.NewManualTimeSetterLinter)
		if err := linter.Input(data, name); err != nil {
			t.Fatalf("failed to Input: %s", err)
		}
		if errs := linter.Errors(); len(errs) > 0 {
			t.Fatalf("failed to lint: %s", name)
		}
	}

}
