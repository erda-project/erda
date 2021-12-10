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

var CONFIGSETPLATFORM = command.Command{
	Name:       "set-platform",
	ParentName: "CONFIG",
	ShortHelp:  "set platform in config file for Erda CLI",
	Example:    "$ erda-cli config set-platform <name> [flags]",
	Args: []command.Arg{
		command.StringArg{}.Name("name"),
	},
	Flags: []command.Flag{
		command.StringFlag{Short: "", Name: "server", Doc: "the http endpoint for openapi of platform", DefaultValue: "https://openapi.erda.cloud"},
		command.StringFlag{Short: "", Name: "org", Doc: "an org under the platform", DefaultValue: ""},
	},
	Run: ConfigOpsWSetPlatform,
}

func ConfigOpsWSetPlatform(ctx *command.Context, name, server, org string) error {
	err := configOpsW("set-platform", name, server, org, "")
	if err != nil {
		return err
	}

	ctx.Succ(fmt.Sprintf("Platform \"%s\" set.", name))
	return nil
}
