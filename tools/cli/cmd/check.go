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

	"github.com/erda-project/erda/pkg/footnote"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

// CHECK command
var CHECK = command.Command{
	Name:      "check",
	ShortHelp: "Validate dice.yml file",
	Example: `
  $ dice check -f dice.yml
`,
	Flags: []command.Flag{
		command.StringFlag{Short: "f", Name: "file",
			Doc: "Specify the path of dice.yml file, default: .dice/dice.yml", DefaultValue: ""},
	},
	Run:    RunCheck,
	Hidden: true,
}

// RunCheck validates dice.yml file
func RunCheck(ctx *command.Context, ymlPath string) error {
	var yml []byte
	var err error
	if ymlPath != "" {
		yml, err = format.ReadYml(ymlPath)
		if err != nil {
			return err
		}
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
	dyml, err := diceyml.New(yml, false)
	if err != nil {
		return err
	}
	verr := dyml.Validate()
	if len(verr) == 0 {
		ctx.Succ("OK")
		return nil
	}
	fnote := footnote.New(string(yml))
	for regex, note := range verr {
		fnote.NoteRegex(regex, note.Error())
	}
	fmt.Printf("%+v\n", fnote.Dump())
	return nil
}
