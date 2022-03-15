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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

var CONFIGSET = command.Command{
	Name:       "set",
	ParentName: "CONFIG",
	ShortHelp:  "set project workspace config",
	Example: `
  $ erda-cli config set HPA=enable --orgName xxx --projectName yyy --workspace DEV
`,
	Args: []command.Arg{
		command.StringArg{}.Name("feature"),
	},
	Flags: []command.Flag{
		command.StringFlag{Name: "orgName", Doc: "[required]which org the project belongs", DefaultValue: ""},
		command.StringFlag{Name: "projectName", Doc: "[required]which project's feature to delete", DefaultValue: ""},
		command.StringFlag{Name: "workspace", Doc: "[optional]which workspace's feature to delete", DefaultValue: ""},
	},
	Run: RunFeaturesSet,
}

func RunFeaturesSet(ctx *command.Context, feature, orgName, projectName, workspace string) error {
	if projectName == "" || workspace == "" || orgName == "" {
		return fmt.Errorf(
			utils.FormatErrMsg("config set", "failed to set config, one of the flags [orgName, projectName, workspace] not set", true))
	}

	if err := SetProjectWorkspaceFeatures(ctx, feature, orgName, projectName, workspace); err != nil {
		ctx.Error("failed to set config: " + err.Error() + "\n")
		return fmt.Errorf(utils.FormatErrMsg("config set", "failed to set configs: "+err.Error(), true))
	}

	ctx.Succ("config set success\n")
	return nil
}

// SetProjectWorkspaceFeatures 为指定项目的指定环境设置 features
func SetProjectWorkspaceFeatures(ctx *command.Context, feature, orgName, projectName, workspace string) error {
	var resp apistructs.ProjectWorkSpaceAbilityResponse

	uop, err := GetUserOrgProjID(ctx, orgName, projectName)
	if err != nil {
		return errors.New("can not get orgID or userID or projectID: " + err.Error())
	}

	prjId, _ := strconv.ParseUint(uop.ProjectId, 10, 64)
	abilities, err := ParseProjectWorkspaceFeatures(feature)
	if err != nil {
		return errors.New("parse config list failed: " + err.Error())
	}

	req := apistructs.ProjectWorkSpaceAbility{
		ProjectID: prjId,
		Workspace: workspace,
		Abilities: abilities,
	}

	response, err := ctx.Post().Path("/api/project-workspace-abilities").
		Header(httputil.OrgHeader, uop.OrgId).
		JSONBody(req).Do().JSON(&resp)

	if err != nil {
		return errors.New("failed to request set project workspace config: " + err.Error())
	}

	if !response.IsOK() {
		return errors.New(fmt.Sprintf("ailed to request set project workspace configs, status-code: %d, content-type: %s, raw bod: %s",
			response.StatusCode(), response.ResponseHeader("Content-Type"), string(response.Body())))
	}

	if !resp.Success {
		return errors.New(fmt.Sprintf("failed to request, error code: %s, raw bod: %s, error message: %s", resp.Error.Code, string(response.Body()), resp.Error.Msg))
	}
	return nil
}

// ParseProjectWorkspaceFeatures 解析 project workspace 的 feature list
func ParseProjectWorkspaceFeatures(featureList string) (string, error) {
	features := make(map[string]string)

	fts := strings.Split(featureList, ",")

	if len(fts) > 0 {
		for _, f := range fts {
			kv := strings.Split(f, "=")
			switch {
			case len(kv) > 2:
				return "", errors.New(fmt.Sprintf("set configs with invalid parameter %s", f))
			case len(kv) == 2:
				features[kv[0]] = kv[1]
			case len(kv) == 1:
				features[kv[0]] = "enable"
			}
		}
	}

	featureStr, err := json.Marshal(features)
	if err != nil {
		return "", errors.New(fmt.Sprintf("set config failed to Marshal request, error: %v ", err))
	}
	return string(featureStr), nil
}
