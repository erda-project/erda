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
	"github.com/erda-project/erda/tools/cli/command"
)

var CONFIGCURCTX = command.Command{
	Name:       "current-context",
	ParentName: "CONFIG",
	ShortHelp:  "show current context set in config file for Erda CLI",
	Example:    "$ erda-cli config current-context",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "if true, don't print headers (default print headers)", DefaultValue: false},
	},
	Run: ConfigCurCtx,
}

func ConfigCurCtx(ctx *command.Context, noHeaders bool) error {
	return configOps("current-context", noHeaders)
}
