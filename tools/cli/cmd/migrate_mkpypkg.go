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
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/database/sqlparser/pygrator"
	"github.com/erda-project/erda/tools/cli/command"
)

var MigrateMkPyPkg = command.Command{
	ParentName:     "Migrate",
	Name:           "mkpypkg",
	ShortHelp:      "generate python package",
	LongHelp:       "generate python package",
	Example:        "erda-cli migrate mkpypkg --filename=my_script.py --mysql-host localhost --mysql-port 3306 --mysql-username root --mysql-password *** --database erda --sandbox-port 3307",
	Hidden:         true,
	DontHideCursor: false,
	Args:           nil,
	Flags: append(
		mysqlFlags,
		command.StringFlag{
			Short:        "",
			Name:         "filename",
			Doc:          "python script filename",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "",
			Name:         "requirements",
			Doc:          "requirements.txt file name",
			DefaultValue: "",
		},
		command.BoolFlag{
			Short:        "",
			Name:         "commit",
			Doc:          "",
			DefaultValue: false,
		},
	),
	Run: RunMigrateMkPyPkg,
}

type textFile struct {
	name string
	data []byte
}

func (t textFile) GetName() string {
	return t.name
}

func (t textFile) GetData() []byte {
	return t.data
}

func RunMigrateMkPyPkg(ctx *command.Context, host string, port int, username, password, database string, sandboxPort int,
	filename, requirements string, commit bool) (err error) {
	logrus.Infoln("Erda Migrator generate the python package")

	var developerScript = textFile{
		name: filename,
		data: nil,
	}

	developerScript.data, err = ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	var p = pygrator.Package{
		DeveloperScript: developerScript,
		Requirements:    nil,
		Settings: pygrator.Settings{
			Engine:   pygrator.DjangoMySQLEngine,
			User:     username,
			Password: password,
			Host:     host,
			Port:     port,
			Name:     database,
			TimeZone: pygrator.TimeZoneAsiaShanghai,
		},
		Commit: commit,
	}

	p.Requirements, err = ioutil.ReadFile(requirements)
	if err != nil {
		logrus.WithError(err).Warnln("failed to read requirements.txt, use default")
		p.Requirements = []byte(pygrator.BaseRequirements)
	}

	return p.Make()
}
