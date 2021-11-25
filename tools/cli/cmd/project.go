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

	"github.com/erda-project/erda/tools/cli/dicedir"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var PROJECT = command.Command{
	Name:      "project",
	ShortHelp: "List projects",
	Example:   "erda-cli project",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization ", DefaultValue: 0},
		command.IntFlag{Short: "", Name: "page-size", Doc: "the number of page size", DefaultValue: 10},
	},
	Run: GetProjects,
}

func GetProjects(ctx *command.Context, noHeaders bool, orgId uint64, pageSize int) error {
	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return errors.New("Invalid organization id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = ctx.CurrentOrg.ID
	}

	num := 0
	err := dicedir.PagingView(func(pageNo, pageSize int) (bool, error) {
		pagingProject, err := common.GetPagingProjectList(ctx, orgId, pageNo, pageSize)
		if err != nil {
			return false, err
		}

		data := [][]string{}
		for _, p := range pagingProject.List {
			data = append(data, []string{
				strconv.FormatUint(p.ID, 10),
				p.Name,
				p.DisplayName,
				p.Desc,
			})
		}

		t := table.NewTable()
		if !noHeaders {
			t.Header([]string{
				"ProjectID", "Name", "DisplayName", "Description",
			})
		}
		err = t.Data(data).Flush()
		if err != nil {
			return false, err
		}

		num += len(pagingProject.List)
		return pagingProject.Total > num, nil
	}, "Continue to display project?", pageSize)
	if err != nil {
		return err
	}

	return nil
}
