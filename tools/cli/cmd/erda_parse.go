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
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/erda-project/erda/tools/cli/prettyjson"
)

var PARSE = command.Command{
	Name:       "parse",
	ParentName: "ERDA",
	ShortHelp:  "Parse erda.yml",
	Example:    "$ erda-cli erda parse -f .erda/erda.yml",
	Flags: []command.Flag{
		command.StringFlag{"f", "file",
			"Specify the path of erda.yml file, default: .erda/erda.yml", ""},
		command.StringFlag{"s", "str", "Provide the content of erda.yml as a string", ""},
		command.BoolFlag{"", "dev", "If true, parse the erda.yml file in development environment ", false},
		command.BoolFlag{"", "test", "If true, parse the erda.yml file in test environment", false},
		command.BoolFlag{"", "staging", "If true, parse the erda.yml file in staging environment", false},
		command.BoolFlag{"", "prod", "If true, parse the erda.yml in production environment", false},
		command.StringFlag{"o", "output", "Output format. One of yaml|json", "yaml"},
	},
	Run:    RunParse,
	Hidden: true,
}

func RunParse(ctx *command.Context, ymlPath string, ymlContent string, dev, test, staging, prod bool, output string) error {
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
		ymlPath, err = ctx.ErdaYml(true)
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
	switch strings.ToLower(output) {
	case "json":
		jsR, err := dyml.JSON()
		if err != nil {
			return err
		}
		pJSR, err := prettyjson.Format([]byte(jsR))
		res = string(pJSR)
	case "yaml":
		res, err = dyml.YAML()
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Invalid output format %s", output))

	}
	os.Stdout.WriteString(res)
	return nil
}
