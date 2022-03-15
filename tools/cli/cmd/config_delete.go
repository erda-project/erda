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

var CONFIGDELETE = command.Command{
	Name:       "delete",
	ParentName: "CONFIG",
	ShortHelp:  "delete project workspace config",
	Example: `
  $ erda-cli config delete --orgName xxx --projectName yyy  --workspace DEV
  $ erda-cli config delete --orgName xxx --projectName yyy
`,
	Flags: []command.Flag{
		command.StringFlag{Name: "orgName", Doc: "[required]which org the project belongs", DefaultValue: ""},
		command.StringFlag{Name: "projectName", Doc: "[required]which project's feature to delete", DefaultValue: ""},
		command.StringFlag{Name: "workspace", Doc: "[optional]which workspace's feature to delete", DefaultValue: ""},
	},
	Run: RunFeaturesDelete,
}

func RunFeaturesDelete(ctx *command.Context, orgName, projectName, workspace string) error {
	var resp apistructs.ExtensionVersionGetResponse
	var b bytes.Buffer

	if projectName == "" || workspace == "" || orgName == "" {
		return fmt.Errorf(
			utils.FormatErrMsg("config delete", "failed to delete config, one of the flags [orgName, projectName, workspace] not set", true))
	}

	uop, err := GetUserOrgProjID(ctx, orgName, projectName)
	if err != nil {
		return fmt.Errorf(
			utils.FormatErrMsg("config delete", "failed to delete config, can not get orgID or userID or projectID: "+err.Error(), true))
	}

	urlPath := ""
	if workspace != "" {
		urlPath = fmt.Sprintf("/api/project-workspace-abilities?projectID=%s&workspace=%s", uop.ProjectId, workspace)
	} else {
		urlPath = fmt.Sprintf("/api/project-workspace-abilities?projectID=%s", uop.ProjectId)
	}

	response, err := ctx.Delete().Path(urlPath).
		Header(httputil.OrgHeader, uop.OrgId).
		Do().Body(&b)

	if err != nil {
		return fmt.Errorf(
			utils.FormatErrMsg("config delete", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return fmt.Errorf(utils.FormatErrMsg("config delete",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err = json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf(utils.FormatErrMsg("config delete",
			fmt.Sprintf("failed to unmarshal config delete response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return fmt.Errorf(utils.FormatErrMsg("config delete",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	ctx.Succ("config delete success\n")
	return nil
}
