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
