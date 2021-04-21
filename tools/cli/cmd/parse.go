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

	"github.com/erda-project/erda/pkg/parser/diceyml"
	. "github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

var PARSE = Command{
	Name:      "parse",
	ShortHelp: "Parse the dice.yml file",
	Flags: []Flag{
		StringFlag{"f", "file",
			"Specify the path of dice.yml file, default: .dice/dice.yml", ""},
		StringFlag{"s", "str", "Provide the content of dice.yml file as a string", ""},
		BoolFlag{"", "dev", "Parse the dice.yml file in development environment ", false},
		BoolFlag{"", "test", "Parse the dice.yml file in test environment", false},
		BoolFlag{"", "staging", "Parse the dice.yml file in staging environment", false},
		BoolFlag{"", "prod", "Parse the dice.yml in production environment", false},
		BoolFlag{"o", "output", "Output the content as yaml", false},
	},
	Run:    RunParse,
	Hidden: true,
}

func RunParse(ctx *Context, ymlPath string, ymlContent string, dev, test, staging, prod, outputYml bool) error {
	var yml []byte
	var err error
	if ymlPath != "" {
		yml, err = format.ReadYml(ymlPath)
		if err != nil {
			return err
		}
	} else if ymlContent != "" {
		yml = []byte(ymlContent)
	} else {
		ymlPath, err = ctx.DiceYml(true)
		if err != nil {
			return err
		}
		yml, err = format.ReadYml(ymlPath)
		if err != nil {
			return err
		}
	}
	dyml, err := diceyml.New(yml, true)
	if err != nil {
		return err
	}
	if dev {
		if err := dyml.MergeEnv("development"); err != nil {
			return err
		}
	} else if test {
		if err := dyml.MergeEnv("test"); err != nil {
			return err
		}
	} else if staging {
		if err := dyml.MergeEnv("staging"); err != nil {
			return err
		}
	} else if prod {
		if err := dyml.MergeEnv("production"); err != nil {
			return err
		}
	}
	var res string
	if outputYml {
		res, err = dyml.YAML()
		if err != nil {
			return err
		}
	} else {
		res, err = dyml.JSON()
		if err != nil {
			return err
		}
	}
	os.Stdout.WriteString(res)
	return nil
}
