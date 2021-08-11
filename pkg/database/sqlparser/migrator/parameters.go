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
	"fmt"
	"net/url"
	"time"
)

type Parameters interface {
	ScriptsParameters

	// MySQLParameters returns MySQL DSN configuration
	MySQLParameters() *DSNParameters

	// SandboxParameters returns Sandbox DSN configuration
	SandboxParameters() *DSNParameters

	// DebugSQL gets weather to debug SQL executing
	DebugSQL() bool

	SkipMigrationLint() bool
	SkipSandbox() bool
	SkipPreMigrate() bool
	SkipMigrate() bool
}

type DSNParameters struct {
	Username  string
	Password  string
	Host      string
	Port      int
	Database  string
	ParseTime bool
	Timeout   time.Duration
}

func (c DSNParameters) Format(database bool) (dsn string) {
	dsnPat := "%s:%s@tcp(%s:%v)/%s"
	if database {
		dsn = fmt.Sprintf(dsnPat, c.Username, c.Password, c.Host, c.Port, c.Database)
	} else {
		dsn = fmt.Sprintf(dsnPat, c.Username, c.Password, c.Host, c.Port, "")
	}

	var params = make(url.Values)
	if c.ParseTime {
		params.Add("parseTime", "true")
	}
	if c.Timeout == 0 {
		c.Timeout = time.Second * 150
	}
	params.Add("timeout", fmt.Sprintf("%vs", int(c.Timeout.Seconds())))
	params.Add("multiStatements", "true")
	params.Add("charset", "utf8mb4,utf8")
	return dsn + "?" + params.Encode()
}
