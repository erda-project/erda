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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var PROJECTCLEAR = command.Command{
	Name:       "clear",
	ParentName: "PROJECT",
	ShortHelp:  "Clear project by delete runtimes and addons",
	Example:    "$ erda-cli project clear --project-id=<id>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "the id of a project", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "the name of a project", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "workspace", Doc: "the env workspace of a project, if set only clear runtimes and addons in the specific workspace", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "wait-runtime", Doc: "the minutes to wait runtimes deleted", DefaultValue: 3},
		command.IntFlag{Short: "", Name: "wait-addon", Doc: "the minutes to wait addons deleted", DefaultValue: 3},
	},
	Run: ClearProject,
}

func ClearProject(ctx *command.Context, orgId, projectId uint64, org, project, workspace string, waitRuntime, waitAddon int) error {
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

	apps, err := common.GetApplications(ctx, orgId, projectId)
	if err != nil {
		return err
	}

	// Clear Runtimes
	appList := []interface{}{}
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
	doTaskWithTimeout(appList, func(id interface{}) bool {
		aId, ok := id.(uint64)
		if !ok {
			return false
		}

		return checkApplication(ctx, orgId, aId, workspace)
	}, time.Duration(waitRuntime)*time.Minute)

	// Clear Addons
	resp, err := common.GetAddonList(ctx, orgId, projectId)
	if err != nil {
		return err
	}
	addonList := []interface{}{}
	for _, a := range resp.Data {
		if workspace != "" && workspace != a.Workspace {
			continue
		}
		addonList = append(addonList, a.ID)
	}

	doTaskWithTimeout(addonList, func(id interface{}) bool {
		aId, ok := id.(string)
		if !ok {
			return false
		}
		err = common.DeleteAddon(ctx, orgId, aId)
		if err == nil {
			return true
		}
		return false
	}, 3*time.Minute)

	// Check Clear Addons Done
	doTaskWithTimeout([]interface{}{projectId}, func(id interface{}) bool {
		pId, ok := id.(uint64)
		if !ok {
			return false
		}

		var as []apistructs.AddonFetchResponseData
		resp, err = common.GetAddonList(ctx, orgId, pId)
		for _, a := range resp.Data {
			if workspace != "" && workspace != a.Workspace {
				continue
			}
			as = append(as, a)
		}
		if err == nil && len(as) == 0 {
			return true
		}
		return false
	}, time.Duration(waitAddon)*time.Minute)

	ctx.Succ("Project clear success.")
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

type taskRunner func(id interface{}) bool

func doTaskWithTimeout(ids []interface{}, c taskRunner, timeout time.Duration) error {
	wg := sync.WaitGroup{}
	timeoutCtx, _ := context.WithTimeout(context.Background(), timeout)

	for _, id := range ids {
		wg.Add(1)
		go func(id interface{}) {
			defer wg.Done()
			timeTicker := time.NewTicker(2 * time.Second)
			for {
				select {
				case <-timeTicker.C:
					if c(id) {
						return
					}
				case <-timeoutCtx.Done():
					return
				}
			}
		}(id)
	}
	wg.Wait()
	if timeoutCtx.Err() != nil {
		return timeoutCtx.Err()
	}
	return nil
}
