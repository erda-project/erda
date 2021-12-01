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
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/dicedir"

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ORGMEMBER = command.Command{
	Name:       "member",
	ParentName: "ORG",
	ShortHelp:  "Display members of the organization",
	Example:    "$ erda-cli org member --org=<name>",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "When using the default or custom-column output format, don't print headers (default print headers)", DefaultValue: false},
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "page-size", Doc: "The number of page size", DefaultValue: 10},
		command.StringListFlag{Short: "", Name: "roles", Doc: "The roles to list", DefaultValue: nil},
	},
	Run: OrgMember,
}

func OrgMember(ctx *command.Context, noHeaders bool, orgId uint64, org string, pageSize int, roles []string) error {
	checkOrgParam(org, orgId)

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	num := 0
	dicedir.PagingView(func(pageNo, pageSize int) (bool, error) {
		pagingMembers, err := common.GetPagingMembers(ctx, apistructs.OrgScope, orgId, roles, pageNo, pageSize)
		if err != nil {
			fmt.Println(err)
			return false, err
		}

		data := [][]string{}
		for _, m := range pagingMembers.List {
			data = append(data, []string{
				m.Nick,
				m.Name,
				m.Email,
				m.Mobile,
				strings.Join(m.Roles, ","),
			})
		}

		t := table.NewTable()
		if !noHeaders {
			t.Header([]string{
				"Nick", "Name", "Email", "Mobile", "Roles",
			})
		}
		err = t.Data(data).Flush()
		if err != nil {
			return false, err
		}
		num += len(pagingMembers.List)
		return pagingMembers.Total > num, nil
	}, "Continue to display members?", pageSize, command.Interactive)

	return nil
}
