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

	"github.com/erda-project/erda/tools/cli/command"
)

var CONFIGDELETECTX = command.Command{
	Name:       "delete-context",
	ParentName: "CONFIG",
	ShortHelp:  "delete context in config file for Erda CLI",
	Example:    "$ erda-cli config delete-context <name>",
	Args: []command.Arg{
		command.StringArg{}.Name("name"),
	},
	Run: ConfigOpsWDeleteCtx,
}

func ConfigOpsWDeleteCtx(ctx *command.Context, name string) error {
	err := configOpsW("delete-context", name, "", "", "")
	if err != nil {
		return err
	}

	ctx.Succ(fmt.Sprintf("Deleted context \"%s\".", name))
	return nil
}
