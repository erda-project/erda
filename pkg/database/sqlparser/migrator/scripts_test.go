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

package migrator_test

import (
	"path/filepath"
	"testing"

	"github.com/erda-project/erda/pkg/database/sqllint/rules"
	"github.com/erda-project/erda/pkg/database/sqlparser/migrator"
)

type parameter struct {
	workdir      string
	migrationDir string
}

func (p parameter) MySQLParameters() *migrator.DSNParameters {
	return nil
}

func (p parameter) SandboxParameters() *migrator.DSNParameters {
	return nil
}

func (p parameter) Database() string {
	return ""
}

func (p parameter) MigrationDir() string {
	return p.migrationDir
}

func (p parameter) Modules() []string {
	return nil
}

func (p parameter) Workdir() string {
	return p.workdir
}

func (p parameter) DebugSQL() bool {
	return false
}

func (p parameter) NeedErdaMySQLLint() bool {
	return false
}

func (p parameter) Rules() []rules.Ruler {
	return nil
}

func TestNewScripts(t *testing.T) {
	var p = parameter{
		workdir:      "..",
		migrationDir: "testdata/new_scripts_test_data",
	}
	var (
		serviceB      = "service_b"
		serviceBBase0 = "testdata/new_scripts_test_data/service_b/this_is_base.sql"
	)

	scripts, err := migrator.NewScripts(p)
	if err != nil {
		t.Fatal(err)
	}

	for name, service := range scripts.Services {
		t.Log(name)
		for _, filename := range service.Filenames() {
			t.Log("\t", filename)
		}
	}

	// assert
	if first := scripts.Services[serviceB].Filenames()[0]; first != filepath.Base(serviceBBase0) {
		t.Fatal("base file error")
	}
}

func TestNewScripts2(t *testing.T) {
	var p = parameter{
		workdir:      "..",
		migrationDir: "testdata/new_scripts_test_data2",
	}

	_, err := migrator.NewScripts(p)
	if err == nil {
		t.Fatal("fails")
	}
}

func TestScripts_AlterPermissionLint(t *testing.T) {
	var p = parameter{
		workdir:      "..",
		migrationDir: "testdata/alter_permission_test_data",
	}

	scripts, err := migrator.NewScripts(p)
	if err != nil {
		t.Fatal(err)
	}

	// assert: if AlterPermissionLint returns err == nil, test fails
	if err = scripts.AlterPermissionLint(); err == nil {
		t.Fatal(err)
	} else {
		t.Log(err)
	}
}
