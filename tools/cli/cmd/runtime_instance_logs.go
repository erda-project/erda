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

var RUNTIMEINSTANCELOGS = command.Command{
	ParentName: "RUNTIMEINSTANCE",
	Name:       "logs",
	ShortHelp:  "View runtime instance logs",
	Example: ` $ erda-cli runtime instance logs pod-0
  $ erda-cli runtime instance logs --service web pod-0 --watch
  $ erda-cli runtime instance logs --service web --instance pod-0 --watch`,
	Args: []command.Arg{
		command.StringArg{}.Name("instance").Option(),
	},
	Flags: []command.Flag{
		command.StringFlag{Short: "", Name: "workspace", Doc: "workspace to query", DefaultValue: ""},
		command.Uint64Flag{Short: "i", Name: "runtime-id", Doc: "show instances for a specific runtime id", DefaultValue: 0},
		command.StringFlag{Short: "s", Name: "service", Doc: "filter by service name", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "instance", Doc: "instance name", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "tail", Doc: "number of recent log lines to fetch first", DefaultValue: 200},
		command.StringFlag{Short: "", Name: "stream", Doc: "log stream: stdout, stderr, or all", DefaultValue: ""},
		command.BoolFlag{Short: "w", Name: "watch", Doc: "watch for new log lines", DefaultValue: false},
	},
	Run: RuntimeInstanceLogs,
}
