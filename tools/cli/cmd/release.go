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
	"github.com/erda-project/erda/tools/cli/dicedir"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var RELEASE = command.Command{
	Name:      "release",
	ShortHelp: "List releases",
	Example:   "$ erda-cli release",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "application-id", Doc: "The id of an application", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "The name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "branch", Doc: "The branch of an application", DefaultValue: ""},
	},
	Run: ReleaseList,
}

func ReleaseList(ctx *command.Context, noHeaders bool, orgId, applicationId uint64, org, branch string) error {
	checkOrgParam(org, orgId)

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	if applicationId <= 0 {
		return errors.New("Invalid application id")
	}

	num := 0
	err = dicedir.PagingView(func(pageNo, pageSize int) (bool, error) {
		list, err := common.GetPagingReleases(ctx, orgId, applicationId, branch, pageNo, pageSize)
		if err != nil {
			return false, err
		}

		data := [][]string{}
		for _, l := range list.Releases {
			data = append(data, []string{
				l.ReleaseID,
				l.ReleaseName,
				l.CreatedAt.String(),
			})
		}

		t := table.NewTable()
		if !noHeaders {
			t.Header([]string{
				"ReleaseD", "ReleaseName", "CreateAt",
			})
		}
		err = t.Data(data).Flush()
		if err != nil {
			return false, err
		}

		num += len(list.Releases)

		return int(list.Total) > num, nil
	}, "Continue to display releases?", 10, command.Interactive)
	if err != nil {
		return err
	}

	return nil
}
