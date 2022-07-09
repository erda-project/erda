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

var Gw = command.Command{
	ParentName:                 "",
	Name:                       "gw",
	ShortHelp:                  "Erda Gateway Tools",
	LongHelp:                   "erda-cli gw --help",
	Example:                    "erda-cli gw ls --cluster erda-cloud -o xxxx",
	Hidden:                     false,
	DontHideCursor:             false,
	Args:                       nil,
	MarkFlagRequired:           nil,
	RegisterFlagCompletionFunc: nil,
	ValidArgsFunction:          nil,
	Flags:                      nil,
	Run:                        nil,
}
