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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

var CONFIGUPDATE = command.Command{
	Name:       "update",
	ParentName: "CONFIG",
	ShortHelp:  "update project workspace config",
	Example: `
  $ erda-cli config update ECI=disable --orgName xxx --projectName yyy --workspace DEV
`,
	Args: []command.Arg{
		command.StringArg{}.Name("feature"),
	},
	Flags: []command.Flag{
		command.StringFlag{Name: "orgName", Doc: "[required]which org the project belongs", DefaultValue: ""},
		command.StringFlag{Name: "projectName", Doc: "[required]which project's feature to delete", DefaultValue: ""},
		command.StringFlag{Name: "workspace", Doc: "[optional]which workspace's feature to delete", DefaultValue: ""},
	},
	Run: RunFeaturesUpdate,
}

func RunFeaturesUpdate(ctx *command.Context, feature, orgName, projectName, workspace string) error {
	var resp apistructs.ProjectWorkSpaceAbilityResponse

	if projectName == "" || workspace == "" || orgName == "" {
		return fmt.Errorf(
			utils.FormatErrMsg("fconfig update", "failed to update config, one of the flags [orgName, projectName, workspace] not set", true))
	}

	uop, err := GetUserOrgProjID(ctx, orgName, projectName)
	if err != nil {
		return fmt.Errorf(
			utils.FormatErrMsg("config update", "failed to update config, can not get orgID or userID or projectID: "+err.Error(), true))
	}

	prjId, _ := strconv.ParseUint(uop.ProjectId, 10, 64)
	abilities, err := ParseProjectWorkspaceFeatures(feature)
	if err != nil {
		return fmt.Errorf(
			utils.FormatErrMsg("config update", fmt.Sprintf("failed to update config, parse config list failed: %v", err), true))
	}

	req := apistructs.ProjectWorkSpaceAbility{
		ProjectID: prjId,
		Workspace: workspace,
		Abilities: abilities,
	}

	response, err := ctx.Put().Path("/api/project-workspace-abilities").
		Header(httputil.OrgHeader, uop.OrgId).
		JSONBody(req).Do().JSON(&resp)
	if err != nil {
		return fmt.Errorf(
			utils.FormatErrMsg("config update", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return fmt.Errorf(utils.FormatErrMsg("config update",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), string(response.Body())), false))
	}

	if !resp.Success {
		return fmt.Errorf(utils.FormatErrMsg("config update",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	ctx.Succ("config update success\n")
	return nil
}
