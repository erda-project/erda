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

var APPLICATION = command.Command{
	Name:      "application",
	ShortHelp: "List applications",
	Example:   "$ erda-cli application --project=<name>",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "The id of a project", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "The name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "The name of a project", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "page-size", Doc: "The number of page size", DefaultValue: 10},
	},
	Run: GetApplications,
}

func GetApplications(ctx *command.Context, noHeaders bool, orgId, projectId uint64, org, project string, pageSize int) error {
	checkOrgParam(org, orgId)
	checkProjectParam(project, projectId)

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	projectId, err = getProjectId(ctx, orgId, project, projectId)
	if err != nil {
		return err
	}

	num := 0
	err = dicedir.PagingView(
		func(pageNo, pageSize int) (bool, error) {
			pagingApplication, err := common.GetPagingApplications(ctx, orgId, projectId, pageNo, pageSize)
			if err != nil {
				return false, err
			}

			data := [][]string{}
			for _, p := range pagingApplication.List {
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
					"ApplicationID", "Name", "DisplayName", "Description",
				})
			}
			err = t.Data(data).Flush()
			if err != nil {
				return false, err
			}

			num += len(pagingApplication.List)
			return pagingApplication.Total > num, nil
		}, "Continue to display applications?", pageSize, command.Interactive)
	if err != nil {
		return err
	}

	return nil
}

func checkApplicationParam(application string, applicationId uint64) {
	if application != "" && applicationId != 0 {
		fmt.Println("Both --application and --application-id are set, we will only use name set by --application")
	}
}

func getApplicationId(ctx *command.Context, orgId, projectId uint64, application string, applicationId uint64) (uint64, error) {
	if application != "" {
		appId, err := common.GetApplicationIdByName(ctx, orgId, projectId, application)
		if err != nil {
			return applicationId, err
		}
		applicationId = appId
	}
	if applicationId < 0 {
		return applicationId, errors.New("Invalid application id")
	}

	return applicationId, nil
}
