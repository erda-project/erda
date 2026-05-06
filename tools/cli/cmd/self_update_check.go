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

var UPDATECHECK = command.Command{
	ParentName: "UPDATE",
	Name:       "check",
	ShortHelp:  "check whether a newer erda-cli version is available",
	Example: ` $ erda-cli update check
  $ erda-cli update check --channel alpha
  $ erda-cli update check --version 2.4.1`,
	Flags: []command.Flag{
		command.StringFlag{Name: "channel", Doc: "release channel: stable, beta, alpha; uses configured default when omitted"},
		command.StringFlag{Name: "version", Doc: "specific version to check"},
	},
	Run: CheckUpdate,
}
