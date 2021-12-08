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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ADDON = command.Command{
	Name:      "addon",
	ShortHelp: "List addons",
	Example:   "$ erda-cli addon --project=<name> --workspace=<ENV>",
	Flags: []command.Flag{
		command.BoolFlag{Short: "", Name: "no-headers", Doc: "If true, don't print headers (default print headers)", DefaultValue: false},
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "The id of a project", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "The name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "The name of a project", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "workspace", Doc: "The env workspace", DefaultValue: ""},
	},
	Run: GetAddons,
}

func GetAddons(ctx *command.Context, noHeaders bool, orgId, projectId uint64, org, project, workspace string) error {
	if workspace != "" {
		if !apistructs.WorkSpace(workspace).Valide() {
			return errors.New(fmt.Sprintf("Invalide workspace %s, should be one in %s",
				workspace, apistructs.WorkSpace("").ValideList()))
		}
	}

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

	list, err := common.GetAddonList(ctx, orgId, projectId)
	if err != nil {
		return err
	}

	data := [][]string{}
	for _, l := range list.Data {
		if workspace != "" && l.Workspace != workspace {
			continue
		}
		data = append(data, []string{
			l.ID,
			l.Name,
			l.AddonName,
			l.Workspace,
			strconv.Itoa(l.Reference),
			l.AddonDisplayName,
		})
	}

	t := table.NewTable()
	if !noHeaders {
		t.Header([]string{
			"AddonID", "Name", "AddonName", "ENV", "Reference", "DisplayName",
		})
	}
	return t.Data(data).Flush()
}
