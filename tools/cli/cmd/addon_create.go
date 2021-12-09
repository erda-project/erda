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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var ADDONCREATE = command.Command{
	Name:       "create",
	ParentName: "ADDON",
	ShortHelp:  "create addon",
	Example:    "$ erda-cli addon create --project=<name> --workspace=<DEV/TEST/STAGING/PROD> --addon-type=<erda/custom> --addon-name=<custom/aliyun-rds>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "the id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "the id of a project", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "the name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "the name of a project", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "workspace", Doc: "the env workspace of an addon", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "addon-type", Doc: "the type of the addon, one of [erda|custom]", DefaultValue: "custom"},
		command.StringFlag{Short: "", Name: "addon-name", Doc: "the name of the addon", DefaultValue: "custom"},
		command.StringFlag{Short: "", Name: "name", Doc: "the name of the addon instance", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "configs", Doc: "the configs of the addon instance in format of key/value. (e.g. --configs='key1=value1,key2=value2')"},
		command.StringFlag{Short: "", Name: "plan", Doc: "the plan of the addon instance", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "wait-addon", Doc: "the minutes to wait erad addon create", DefaultValue: 3},
		// command.StringListFlag{Short: "", Name: "erda-yaml", Doc: "the erda.yml path to add addon", DefaultValue: nil},
	},
	Run: AddonCreate,
}

func AddonCreate(ctx *command.Context, orgId, projectId uint64, org, project, workspace,
	addonType, addonName, name, configs, plan string, waitAddon int) error {
	checkOrgParam(org, orgId)
	checkProjectParam(project, projectId)
	if !apistructs.WorkSpace(workspace).Valide() {
		return errors.New(fmt.Sprintf("Invalide workspace %s, should be one in %s",
			workspace, apistructs.WorkSpace("").ValideList()))
	}
	if name == "" {
		return errors.New("Invalid name for addon instance")
	}

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	projectId, err = getProjectId(ctx, orgId, project, projectId)
	if err != nil {
		return err
	}

	switch addonType {
	case "erda":
		if plan == "" || len(strings.Split(plan, ":")) != 2 {
			return errors.New("Invalid plan for addon instance")
		}

		p, err := common.GetProjectByName(ctx, orgId, project)
		if err != nil {
			return err
		}
		err = common.CreateErdaAddon(ctx, orgId, projectId, p.ClusterConfig[workspace], workspace, name, plan, waitAddon)
		if err != nil {
			return err
		}
	case "custom":
		if configs == "" {
			return errors.New("Invalid configs for addon instance")
		}

		kvmap := map[string]interface{}{}
		kvs := strings.Split(configs, ",")
		for _, kv := range kvs {
			kvl := strings.Split(kv, "=")
			if len(kvl) != 2 || len(kvl[0]) == 0 || len(kvl[1]) == 0 {
				return errors.New("Invalid configs for addon instance")
			}
			kvmap[kvl[0]] = kvl[1]
		}
		common.CreateCustomAddon(ctx, orgId, projectId, workspace, addonName, name, kvmap)
	default:
		return errors.New("Invalid addon type")
	}

	ctx.Succ("Addon created.")

	//if erdaYamls != nil {
	//	for _, ey := range erdaYamls {
	//		//if !strings.HasSuffix(ey, "erda.yml") {
	//		//	return errors.New("--erda-yaml only support erda.yml")
	//		//}
	//
	//		var absEY string
	//		if path.IsAbs(ey) {
	//			absEY = ey
	//		} else {
	//			wd, err := os.Getwd()
	//			if err != nil {
	//				return err
	//			}
	//			absEY = path.Join(wd, ey)
	//		}
	//		_, err = os.Stat(absEY)
	//		if err != nil {
	//			return err
	//		}
	//		yml, err := format.ReadYml(absEY)
	//		if err != nil {
	//			return err
	//		}
	//		dyml, err := diceyml.New(yml, true)
	//		if err != nil {
	//			return err
	//		}
	//
	//		var envType diceyml.EnvType = diceyml.BaseEnv
	//		switch apistructs.WorkSpace(workspace) {
	//		case apistructs.WorkspaceDev:
	//			if err := dyml.MergeEnv("development"); err != nil {
	//				return err
	//			}
	//			envType = diceyml.DevEnv
	//		case apistructs.WorkspaceTest:
	//			if err := dyml.MergeEnv("test"); err != nil {
	//				return err
	//			}
	//			envType = diceyml.TestEnv
	//		case apistructs.WorkspaceStaging:
	//			if err := dyml.MergeEnv("staging"); err != nil {
	//				return err
	//			}
	//			envType = diceyml.StagingEnv
	//		case apistructs.WorkspaceProd:
	//			if err := dyml.MergeEnv("production"); err != nil {
	//				return err
	//			}
	//			envType = diceyml.ProdEnv
	//		}
	//
	//		dyml.
	//			dyml.InsertAddonOptions(envType, name, map[string]string{})
	//
	//		fmt.Println(dyml.JSON())
	//	}
	//}

	return nil
}
