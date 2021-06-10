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

package cmd

import (
	"github.com/erda-project/erda/pkg/envconf"
	"github.com/erda-project/erda/tools/cli/command"
)

var mysqlFlagsEnv_ *mysqlFlagsEnv

type mysqlFlagsEnv struct {
	Host string `env:"ERDA_MYSQL_HOST"`
	Port int `env:"ERDA_MYSQL_PORT"`
	Username string `env:"ERDA_MYSQL_USERNAME"`
	Password string`env:"ERDA_MYSQL_PASSWORD"`
	Database string`env:"ERDA_MYSQL_DATABASE"`
}

func getMySQLFlagsEnvs() *mysqlFlagsEnv {
	if mysqlFlagsEnv_ != nil {
		return mysqlFlagsEnv_
	}
	mysqlFlagsEnv_ = new(mysqlFlagsEnv)
	_ = envconf.Load(mysqlFlags)
	return mysqlFlagsEnv_
}

var mysqlFlags = []command.Flag{
	command.StringFlag{
		Short:        "h",
		Name:         "host",
		Doc:          "[MySQL] connect to host",
		DefaultValue: getMySQLFlagsEnvs().Host,
	},
	command.IntFlag{
		Short:        "P",
		Name:         "port",
		Doc:          "[MySQL] port number to use for connection",
		DefaultValue: getMySQLFlagsEnvs().Port,
	},
	command.StringFlag{
		Short:        "u",
		Name:         "username",
		Doc:          "[MySQl] user for login",
		DefaultValue: getMySQLFlagsEnvs().Username,
	},
	command.StringFlag{
		Short:        "p",
		Name:         "password",
		Doc:          "[MySQL] password to use then connecting to server",
		DefaultValue: getMySQLFlagsEnvs().Password,
	},
	command.StringFlag{
		Short:        "D",
		Name:         "database",
		Doc:          "[MySQL] database to use",
		DefaultValue: getMySQLFlagsEnvs().Database,
	},
}

var Migrate = command.Command{
	ParentName:     "",
	Name:           "migrate",
	ShortHelp:      "Erda MySQL Migrate",
	LongHelp:       "erda-cli migrate --input=.  -h localhost -P 3306 -u root -p mypassword",
	Example:        "erda-cli migrate --input=.  -h localhost -P 3306 -u root -p mypassword",
	Hidden:         false,
	DontHideCursor: false,
	Args:           nil,
	Flags: append(mysqlFlags,
		command.StringFlag{
			Short:        "i",
			Name:         "input",
			Doc:          "migration directory",
			DefaultValue: ".",
		},
		command.StringFlag{
			Short:        "",
			Name:         "lintConfig",
			Doc:          "Erda MySQL Lint config file",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "",
			Name:         "skipLint",
			Doc:          "true: don't do Erda MySQL Lint",
			DefaultValue: false,
		},
	),
	Run: func() {
		panic("not implement")
	},
}

func RunMigrate(ctx *command.Context, input, lintConfig string, skipLint )  {
	
}