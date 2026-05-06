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

var UPDATELIST = command.Command{
	ParentName: "UPDATE",
	Name:       "list",
	ShortHelp:  "list recent remote versions in a channel",
	Example: ` $ erda-cli update list
  $ erda-cli update list --channel alpha`,
	Flags: []command.Flag{
		command.StringFlag{Name: "channel", Doc: "release channel: stable, beta, alpha; uses configured default when omitted"},
	},
	Run: UpdateList,
}
