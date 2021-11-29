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

var APPLICATIONINSPECT = command.Command{
	Name:       "inspect",
	ParentName: "APPLICATION",
	ShortHelp:  "Inspect application",
	Example:    "$ erda-cli application inspect --project=<name>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization ", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "the id of a project ", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "application-id", Doc: "the id of an application ", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "the name of a project ", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "application", Doc: "the name of an application ", DefaultValue: ""},
	},
	Run: ApplicationInspect,
}

func ApplicationInspect(ctx *command.Context, orgId, projectId, applicationId uint64, org, project, application string) error {
	checkOrgParam(org, orgId)
	checkApplicationParam(application, applicationId)

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	projectId, err = getProjectId(ctx, orgId, project, projectId)
	if err != nil {
		return err
	}

	applicationId, err = getApplicationId(ctx, orgId, projectId, application, applicationId)
	if err != nil {
		return err
	}

	resp, err := common.GetApplicationDetail(ctx, orgId, projectId, applicationId)
	if err != nil {
		return err
	}

	s, err := prettyjson.Marshal(resp)
	if err != nil {
		return fmt.Errorf(format.FormatErrMsg("application inspect",
			"failed to prettyjson marshal application data ("+err.Error()+")", false))
	}

	fmt.Println(string(s))

	return nil
}
