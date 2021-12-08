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

	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/erda-project/erda/tools/cli/prettyjson"
)

var PROJECTINSPECT = command.Command{
	Name:       "inspect",
	ParentName: "PROJECT",
	ShortHelp:  "Inspect project detail information",
	Example:    "$ erda-cli project inspect --project-id=<id>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "The id of a project", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "The name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "The name of a project", DefaultValue: ""},
	},
	Run: InspectProject,
}

func InspectProject(ctx *command.Context, orgId, projectId uint64, org, project string) error {
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

	resp, err := common.GetProjectDetail(ctx, orgId, projectId)
	if err != nil {
		return err
	}

	s, err := prettyjson.Marshal(resp.Data)
	if err != nil {
		return fmt.Errorf(format.FormatErrMsg("project inspect",
			"failed to prettyjson marshal project data ("+err.Error()+")", false))
	}

	fmt.Println(string(s))

	return nil
}
