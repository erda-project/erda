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
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ERDA = command.Command{
	Name: "erda",
	ShortHelp: "List erda.yaml in .dice/ directory (current repo)",
	Example: "erda-cli erda",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
	},
	Run: GetErdas,
}

func GetErdas(ctx *command.Context, noHeaders bool) error {
	branch, err := common.GetWorkspaceBranch()
	if err != nil {
		return err
	}

	var erdaymls []string

	var data [][]string
	for _, p := range erdaymls {
		data = append(data, []string{
			branch,
			p,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"branch", "erda",
		})
	}
	return t.Data(data).Flush()
}