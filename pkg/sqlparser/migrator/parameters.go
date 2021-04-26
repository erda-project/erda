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
