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

var ORG = command.Command{
	Name:      "org",
	ShortHelp: "list organizations",
	Example:   "$ erda-cli org",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "if true, don't print headers (default print headers)", DefaultValue: false},
		command.IntFlag{Short: "", Name: "page-size", Doc: "the number of page size", DefaultValue: 10},
	},
	Run: GetOrgs,
}

func GetOrgs(ctx *command.Context, noHeaders bool, pageSize int) error {
	num := 0
	err := dicedir.PagingView(
		func(pageNo, pageSize int) (bool, error) {
			pagingOrgs, err := common.GetPagingOrganizations(ctx, pageNo, pageSize)
			if err != nil {
				return false, err
			}

			data := [][]string{}
			for _, o := range pagingOrgs.List {
				data = append(data, []string{
					strconv.FormatUint(o.ID, 10),
					o.Name,
					o.Desc,
				})
			}

			t := table.NewTable()
			if !noHeaders {
				t.Header([]string{
					"OrgID", "Name", "Description",
				})
			}
			err = t.Data(data).Flush()
			if err != nil {
				return false, err
			}

			num += len(pagingOrgs.List)

			return pagingOrgs.Total > num, nil
		}, "Continue to display organizations?", pageSize, command.Interactive)
	if err != nil {
		return err
	}

	return nil
}

func getOrgId(ctx *command.Context, org string, orgId uint64) (uint64, error) {
	if org != "" {
		o, err := common.GetOrgDetail(ctx, org)
		if err != nil {
			return orgId, err
		}
		orgId = o.ID
	}

	if orgId <= 0 && ctx.CurrentOrg.ID <= 0 {
		return orgId, errors.New("Invalid organization id")
	}

	if orgId == 0 && ctx.CurrentOrg.ID > 0 {
		orgId = ctx.CurrentOrg.ID
	}

	return orgId, nil
}

func checkOrgParam(org string, orgId uint64) {
	if org != "" && orgId != 0 {
		fmt.Println("Both --org and --org-id are set, we will only use name set by --org")
	}
}
