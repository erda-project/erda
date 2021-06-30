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
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/tools/cli/command"
)

func getPortFromEnv() int {
	port := os.Getenv("ERDA_MYSQL_PORT")
	i, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return 3306
	}
	return int(i)
}

var mysqlFlags = []command.Flag{
	command.StringFlag{
		Short:        "h",
		Name:         "host",
		Doc:          "[MySQL] connect to host",
		DefaultValue: os.Getenv("ERDA_MYSQL_HOST"),
	},
	command.IntFlag{
		Short:        "P",
		Name:         "port",
		Doc:          "[MySQL] port number to use for connection",
		DefaultValue: getPortFromEnv(),
	},
	command.StringFlag{
		Short:        "u",
		Name:         "username",
		Doc:          "[MySQl] user for login",
		DefaultValue: os.Getenv("ERDA_MYSQL_USERNAME"),
	},
	command.StringFlag{
		Short:        "p",
		Name:         "password",
		Doc:          "[MySQL] password to use then connecting to server",
		DefaultValue: os.Getenv("ERDA_MYSQL_PASSWORD"),
	},
	command.StringFlag{
		Short:        "D",
		Name:         "database",
		Doc:          "[MySQL] database to use",
		DefaultValue: os.Getenv("ERDA_MYSQL_DATABASE"),
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
	Run: nil,
}

func RunMigrate(ctx *command.Context) {
	logrus.Infoln("Erda Migrator is working")
}