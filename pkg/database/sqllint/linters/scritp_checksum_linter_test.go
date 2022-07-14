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

const scriptChecksumLinterConfig = `
- name: ScriptChecksumLinter
  meta:
  - scriptName: stmt-1
    checksum: 6151443700efb4531da7cf49c4f36f69e23ea8bd33ac0205afec069b8ed3c51
  - scriptName: stmt-2
    checksum: 1d5f3b8a262e6f207feb9bd2e03efccc384aecb53ddd94cf09a44f62000d5db0
`

func TestScriptChecksumLinter_LintOnScript(t *testing.T) {
	cfg, err := sqllint.LoadConfig([]byte(scriptChecksumLinterConfig))
	if err != nil {
		t.Fatal("failed to LoadConfig", err)
	}
	var content = []byte("create table t1 (name varchar(225))")
	t.Run("test incorrect checksum", func(t *testing.T) {
		name := "stmt-1"
		linter := sqllint.New(cfg)
		if err = linter.Input("", name, content); err != nil {
			t.Fatal(err)
		}
		lints := linter.Errors()[name].Lints
		if len(lints) != 1 {
			t.Fatalf("expects len(lints): 1, got: %v", len(lints))
		}
		t.Logf("errors: %v", lints)
	})
	t.Run("test correct checksum", func(t *testing.T) {
		name := "stmt-2"
		linter := sqllint.New(cfg)
		if err := linter.Input("", name, content); err != nil {
			t.Fatal(err)
		}
		lints := linter.Errors()[name].Lints
		if len(lints) > 0 {
			t.Logf("errors: %v", lints)
			t.Fatalf("expects len(lints): 0, got: %v", len(lints))
		}
	})
}
