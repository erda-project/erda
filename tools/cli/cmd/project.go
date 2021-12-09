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
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
)

var PROJECT = command.Command{
	Name:      "project",
	ShortHelp: "list projects",
	Example:   "$ erda-cli project",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "if true, don't print headers (default print headers)", DefaultValue: false},
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization ", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization ", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "page-size", Doc: "the number of page size", DefaultValue: 10},
	},
	Run: GetProjects,
}

func GetProjects(ctx *command.Context, noHeaders bool, orgId uint64, org string, pageSize int) error {
	checkOrgParam(org, orgId)
	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	num := 0
	err = dicedir.PagingView(func(pageNo, pageSize int) (bool, error) {
		pagingProject, err := common.GetPagingProjects(ctx, orgId, pageNo, pageSize)
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

		headers := []string{
			"ProjectID", "Name", "DisplayName", "Description",
		}

		t := table.NewTable()
		if !noHeaders {
			t.Header(headers)
		}
		err = t.Data(data).Flush()
		if err != nil {
			return false, err
		}

		num += len(pagingProject.List)
		return pagingProject.Total > num, nil
	}, "Continue to display project?", pageSize, command.Interactive)
	if err != nil {
		return err
	}

	return nil
}

func getProjectId(ctx *command.Context, orgId uint64, project string, projectId uint64) (uint64, error) {
	if projectId == 0 && project == "" {
		return projectId, errors.New("Must set one of --project or --project-id")
	}

	if project != "" {
		pId, err := common.GetProjectIdByName(ctx, orgId, project)
		if err != nil {
			return projectId, err
		}
		projectId = pId
	}

	if projectId <= 0 {
		return projectId, errors.New("Invalid project id")
	}

	return projectId, nil
}

func checkProjectParam(project string, projectId uint64) {
	if project != "" && projectId != 0 {
		fmt.Println("Both --project and --project-id are set, we will only use name set by --project")
	}
}
