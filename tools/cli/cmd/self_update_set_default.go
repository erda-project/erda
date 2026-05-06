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

var UPDATESETDEFAULT = command.Command{
	ParentName: "UPDATE",
	Name:       "set-default",
	ShortHelp:  "set default update channel",
	Example: ` $ erda-cli update set-default stable
  $ erda-cli update set-default alpha`,
	Args: []command.Arg{
		command.StringArg{}.Name("channel").Option(),
	},
	Run: UpdateSetDefault,
}
