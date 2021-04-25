package migrator_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint/rules"
	"github.com/erda-project/erda/pkg/sqlparser/migrator"
)

type parameter struct {
	workdir string
	migrationDir string
}

func (p parameter) DSN() string {
	return ""
}

func (p parameter) SandboxDSN() string {
	return ""
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

	// 断言: 如果第一个脚本不是标记了 base 的脚本, 说明 base 脚本识别失败
	if first := scripts.Services[serviceB].Filenames()[0]; first != serviceBBase0 {
		t.Fatal("base file error")
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
