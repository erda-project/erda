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
)

const scriptNameLinterConfig = `
- name: ScriptNameLinter
  meta:
    patterns:
    - ^202([0-9]{5})-.+(\.sql|\.py)$
`

func TestScriptNameLinter_LintOnScript(t *testing.T) {
	cfg, err := sqllint.LoadConfig([]byte(scriptNameLinterConfig))
	if err != nil {
		t.Fatal(err)
	}
	t.Run("test incorrect script name", func(t *testing.T) {
		for _, name := range []string{
			"20190101-abc-abc.sql",
			"202201101-abc.sql",
			"20220101-abc.java",
		} {
			linter := sqllint.New(cfg)
			if err := linter.Input("", name, []byte("create table t1 (name varchar(225))")); err != nil {
				t.Fatal(err)
			}
			lints := linter.Errors()[name].Lints
			if len(lints) != 1 {
				t.Fatalf("%s, expects len(lints): 1, got: %v", name, len(lints))
			}
			t.Logf("errors: %v", lints)
		}
	})
	t.Run("correct script name", func(t *testing.T) {
		for _, name := range []string{
			"20210101-abc-def.sql",
			"20220101-abc-def.py",
		} {
			linter := sqllint.New(cfg)
			if err := linter.Input("", name, []byte("create table t1 (name varchar(225))")); err != nil {
				t.Fatal(err)
			}
			lints := linter.Errors()[name].Lints
			if len(lints) > 0 {
				t.Logf("errors: %v", lints)
				t.Fatalf("expects len(lints): 0, got: %v", len(lints))
			}
		}
	})
}
