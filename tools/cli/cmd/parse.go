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
