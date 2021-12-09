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
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
)

var PROJECTCLEAR = command.Command{
	Name:       "clear",
	ParentName: "PROJECT",
	ShortHelp:  "clear project by delete runtimes and addons",
	Example:    "$ erda-cli project clear --project-id=<id> --workspace=<ENV>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "the id of a project", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "the name of a project", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "workspace", Doc: "the env workspace of a project, if set only clear runtimes and addons in the specific workspace", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "wait-runtime", Doc: "minutes to wait runtimes deleted", DefaultValue: 3},
		command.IntFlag{Short: "", Name: "wait-addon", Doc: "minutes to wait addons deleted", DefaultValue: 3},
		command.BoolFlag{Short: "", Name: "delete-apps", Doc: "if true, delete all applications", DefaultValue: false},
		command.BoolFlag{Short: "", Name: "delete-custom-addons", Doc: "if true, delete all custom addons", DefaultValue: false},
	},
	Run: ClearProject,
}

func ClearProject(ctx *command.Context, orgId, projectId uint64, org, project, workspace string, waitRuntime, waitAddon int, deleteApps, deleteCAs bool) error {
	if workspace != "" {
		if !apistructs.WorkSpace(workspace).Valide() {
			return errors.New(fmt.Sprintf("Invalide workspace %s, should be one in %s",
				workspace, apistructs.WorkSpace("").ValideList()))
		}
	}
	if workspace != "" && deleteApps {
		return errors.New("Should not both set --workspace and --deleteApps")
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

	err = clearProject(ctx, orgId, projectId, workspace, waitRuntime, waitAddon, deleteApps, deleteCAs)

	ctx.Succ("Project clear success.")
	return nil
}

func clearProject(ctx *command.Context, orgId, projectId uint64, workspace string, waitRuntime, waitAddon int, deleteApps, deleteCAs bool) error {
	apps, err := common.GetApplications(ctx, orgId, projectId)
	if err != nil {
		return err
	}

	// Clear Runtimes
	var appList []uint64
	for _, app := range apps {
		rs, err := common.GetRuntimeList(ctx, orgId, app.ID, "", "")
		if err != nil {
			return err
		}

		deletedNum := 0
		for _, r := range rs {
			if workspace != "" && r.Extra["workspace"] != workspace {
				continue
			}
			err = common.DeleteRuntime(ctx, orgId, r.ID)
			if err != nil {
				return err
			}
			deletedNum++
		}
		if deletedNum > 0 {
			appList = append(appList, app.ID)
		}
	}
	// Check Clear Runtimes Done
	var checkRuntimesRunners []dicedir.TaskRunner
	for _, aId := range appList {
		appId := aId
		checkRuntimesRunners = append(checkRuntimesRunners, func() bool {
			return checkApplication(ctx, orgId, appId, workspace)
		})
	}
	err = dicedir.DoTaskListWithTimeout(time.Duration(waitRuntime)*time.Minute, checkRuntimesRunners)
	if err != nil {
		return err
	}

	// Clear Addons
	resp, err := common.GetAddonList(ctx, orgId, projectId)
	if err != nil {
		return err
	}
	var addonList []string
	for _, a := range resp.Data {
		if workspace != "" && workspace != a.Workspace {
			continue
		}
		if !deleteCAs && a.Category == "custom" {
			continue
		}
		addonList = append(addonList, a.ID)
	}

	var deleteAddonRunners []dicedir.TaskRunner
	for _, aId := range addonList {
		// make local appId
		appId := aId
		deleteAddonRunners = append(deleteAddonRunners,
			func() bool {
				err = common.DeleteAddon(ctx, orgId, appId)
				if err == nil {
					return true
				}
				return false
			})
	}
	err = dicedir.DoTaskListWithTimeout(3*time.Minute, deleteAddonRunners)
	if err != nil {
		return err
	}

	// Check Clear Addons Done
	err = dicedir.DoTaskWithTimeout(func() (bool, error) {
		var as []apistructs.AddonFetchResponseData
		resp, err = common.GetAddonList(ctx, orgId, projectId)
		if err != nil {
			return false, err
		}
		for _, a := range resp.Data {
			if workspace != "" && workspace != a.Workspace {
				continue
			}
			if !deleteCAs && a.Category == "custom" {
				continue
			}
			as = append(as, a)
		}
		if err == nil && len(as) == 0 {
			return true, nil
		}
		return false, err
	}, time.Duration(waitAddon)*time.Minute)
	if err != nil {
		return err
	}

	if deleteApps {
		for _, app := range apps {
			err = common.DeleteApplication(ctx, app.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func checkApplication(ctx *command.Context, orgId uint64, appId uint64, workspace string) bool {
	existNum := 0
	rst, e := common.GetRuntimeList(ctx, orgId, appId, "", "")
	for _, r := range rst {
		if workspace != "" && r.Extra["workspace"] != workspace {
			continue
		}
		existNum++
	}
	if e == nil && existNum == 0 {
		return true
	}

	return false
}
