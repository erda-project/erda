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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

var CONFIGGET = command.Command{
	Name:       "get",
	ParentName: "CONFIG",
	ShortHelp:  "get project workspace config",
	Example: `
  $ erda-cli config get --orgName xxx --projectName yyy --workspace DEV
`,
	Flags: []command.Flag{
		command.StringFlag{Name: "orgName", Doc: "[required]which org the project belongs", DefaultValue: ""},
		command.StringFlag{Name: "projectName", Doc: "[required]which project's feature to delete", DefaultValue: ""},
		command.StringFlag{Name: "workspace", Doc: "[optional]which workspace's feature to delete", DefaultValue: ""},
	},
	Run: RunFeaturesGet,
}

func RunFeaturesGet(ctx *command.Context, orgName, projectName, workspace string) error {
	var resp apistructs.ProjectWorkSpaceAbilityResponse
	var b bytes.Buffer

	if projectName == "" || workspace == "" || orgName == "" {
		return fmt.Errorf(
			utils.FormatErrMsg("config get", "failed to get configd, one of the flags [orgName, projectName, workspace] not set", true))
	}

	uop, err := GetUserOrgProjID(ctx, orgName, projectName)
	if err != nil {
		return fmt.Errorf(
			utils.FormatErrMsg("config get", "failed to get configs, can not get orgID or userID or projectID: "+err.Error(), true))
	}

	urlPath := "/api/project-workspace-abilities/" + uop.ProjectId + "/" + workspace

	response, err := ctx.Get().Path(urlPath).
		Header(httputil.OrgHeader, uop.OrgId).
		Do().Body(&b)

	if err != nil {
		return fmt.Errorf(
			utils.FormatErrMsg("config get", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return fmt.Errorf(utils.FormatErrMsg("config get",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err = json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf(utils.FormatErrMsg("config get",
			fmt.Sprintf("failed to unmarshal get config response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return fmt.Errorf(utils.FormatErrMsg("config get",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	ctx.Info("Configs get result: %s\n", resp.Data.Abilities)
	ctx.Succ("config get success\n")
	return nil
}
