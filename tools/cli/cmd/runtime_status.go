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

var RUNTIMESTATUS = command.Command{
	ParentName: "RUNTIME",
	Name:       "status",
	ShortHelp:  "Show runtime status",
	Example: ` $ erda-cli runtime status
  $ erda-cli runtime status --workspace TEST
  $ erda-cli runtime status --runtime-id <runtimeID>`,
	Flags: []command.Flag{
		command.StringFlag{Short: "w", Name: "workspace", Doc: "workspace to query", DefaultValue: ""},
		command.Uint64Flag{Short: "i", Name: "runtime-id", Doc: "show a specific runtime id", DefaultValue: 0},
	},
	Run: RuntimeStatus,
}
