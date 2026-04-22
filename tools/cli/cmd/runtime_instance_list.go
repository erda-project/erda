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

import "github.com/erda-project/erda/tools/cli/command"

var RUNTIMEINSTANCELIST = command.Command{
	ParentName: "RUNTIMEINSTANCE",
	Name:       "list",
	ShortHelp:  "List runtime instances",
	Example: ` $ erda-cli runtime instance list
  $ erda-cli runtime instance list --service web`,
	Flags: []command.Flag{
		command.StringFlag{Short: "w", Name: "workspace", Doc: "workspace to query", DefaultValue: ""},
		command.Uint64Flag{Short: "i", Name: "runtime-id", Doc: "show instances for a specific runtime id", DefaultValue: 0},
		command.StringFlag{Short: "s", Name: "service", Doc: "filter by service name", DefaultValue: ""},
		command.BoolFlag{Short: "a", Name: "all", Doc: "include stopped instances", DefaultValue: false},
	},
	Run: RuntimeInstanceList,
}
