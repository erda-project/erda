package migrator

import (
	"github.com/erda-project/erda/pkg/sqllint/rules"
)

type Parameters interface {
	// DSN gets MySQL DSN
	DSN() string

	// SandboxDSN gets sandbox DSN
	SandboxDSN() string

	// database name
	Database() string

	// MigrationDir gets migration scripts direction from repo, like .dice/migrations or 4.1/sqls
	MigrationDir() string

	// Modules is the modules for installing.
	// if is nil, to install all modules in the MigrationDir()
	Modules() []string

	// Workdir gets pipeline node workdir
	Workdir() string

	// DebugSQL gets weather to debug SQL executing
	DebugSQL() bool

	NeedErdaMySQLLint() bool

	Rules() []rules.Ruler
}
