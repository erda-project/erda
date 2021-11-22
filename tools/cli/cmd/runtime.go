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
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var RUNTIME = command.Command{
	Name:      "runtime",
	ShortHelp: "List runtimes",
	Example:   "erda-cli runtime",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.IntFlag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "application-id", Doc: "The id of an application", DefaultValue: 0},
	},
	Run: RuntimeList,
}

func RuntimeList(ctx *command.Context, noHeaders bool, orgId, projectId int) error {
	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return errors.New("invalid org id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = int(ctx.CurrentOrg.ID)
	}

	if projectId <= 0 {
		return errors.New("invalid project id")
	}

	list, err := common.GetRuntimeList(ctx, orgId, projectId, "", "")
	if err != nil {
		return err
	}

	data := [][]string{}
	for _, l := range list {
		data = append(data, []string{
			strconv.FormatUint(l.ID, 10),
			l.Name,
			l.CreatedAt.String(),
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"RuntimeID", "Name", "CreateAt",
		})
	}
	return t.Data(data).Flush()
}
