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

var APPLICATION = command.Command{
	Name:      "application",
	ShortHelp: "List applications",
	Example:   "erda-cli application",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.IntFlag{Short: "", Name: "org-id", Doc: "the id of an organization ", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "project-id", Doc: "the id of an project ", DefaultValue: 0},
	},
	Run: GetApplications,
}

func GetApplications(ctx *command.Context, noHeaders bool, orgId, projectId int) error {
	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return errors.New("invalid org id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = int(ctx.CurrentOrg.ID)
	}

	if projectId <= 0 {
		return errors.New("invalid project id")
	}

	list, err := common.GetApplicationList(ctx, orgId, projectId)
	if err != nil {
		return err
	}

	data := [][]string{}
	for i := range list {
		data = append(data, []string{
			strconv.FormatUint(list[i].ID, 10),
			list[i].Name,
			list[i].DisplayName,
			list[i].Desc,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"ApplicationID", "Name", "DisplayName", "Description",
		})
	}
	return t.Data(data).Flush()
}
