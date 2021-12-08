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
	"os"
	"time"

	"github.com/elastic/cloud-on-k8s/pkg/utils/stringsutil"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
)

var PROJECTLOAD = command.Command{
	Name:       "load",
	ParentName: "PROJECT",
	ShortHelp:  "Load project by releases",
	Example:    "$ erda-cli project load --project=<name> --workspace=<ENV> --branch=<b> --version=<v> --config=<filename>",
	Flags: []command.Flag{
		command.Uint64Flag{Short: "", Name: "org-id", Doc: "The id of an organization", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "project-id", Doc: "The id of a project", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "org", Doc: "The name of an organization", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "project", Doc: "The name of the project", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "config", Doc: "The name of the configuration file, specify workspace and version for applications", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "workspace", Doc: "The env workspace to deploy", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "branch", Doc: "If branch is specified, all applications use the same one", DefaultValue: ""},
		command.StringFlag{Short: "", Name: "version", Doc: "If version is specified, all applications use the same one", DefaultValue: ""},
		command.IntFlag{Short: "", Name: "wait-time", Doc: "The number of minutes to wait for runtime creating", DefaultValue: 5},
		command.StringListFlag{Short: "", Name: "skip-apps", Doc: "The applications to skip deploy", DefaultValue: nil},
	},
	Run: ProjectLoad,
}

func ProjectLoad(ctx *command.Context, orgId, projectId uint64, org, project, config, workspace, branch, version string, waitTime int, skipApps []string) error {
	checkOrgParam(org, orgId)
	checkProjectParam(project, projectId)

	if workspace == "" {
		return errors.New("Must specify env workspace")
	} else {
		if !apistructs.WorkSpace(workspace).Valide() {
			return errors.New(fmt.Sprintf("Invalide workspace %s, should be one in %s",
				workspace, apistructs.WorkSpace("").ValideList()))
		}
	}

	if branch == "" {
		return errors.New("Must specify branch")
	}
	if version == "" {
		return errors.New("Must specify version")
	}

	releaseConfig := ReleaseConfig{Applications: make(map[string]ReleaseApp)}
	if config != "" {
		f, err := os.Open(config)
		if err != nil {
			return err
		}
		if err := yaml.NewDecoder(f).Decode(&releaseConfig); err != nil {
			return err
		}
	}

	orgId, err := getOrgId(ctx, org, orgId)
	if err != nil {
		return err
	}

	projectId, err = getProjectId(ctx, orgId, project, projectId)
	if err != nil {
		return err
	}

	appList, err := common.GetApplications(ctx, orgId, projectId)
	if err != nil {
		return err
	}

	var rDeployments []apistructs.RuntimeDeployDTO
	for _, app := range appList {
		if stringsutil.StringInSlice(app.Name, skipApps) {
			continue
		}

		releaseBranch := branch
		releaseVersion := version
		rc, ok := releaseConfig.Applications[app.Name]
		if ok {
			if rc.Branch != "" {
				releaseBranch = rc.Branch
			}
			if rc.Version != "" {
				releaseVersion = rc.Version
			}
		}

		found, r, err := common.ChooseRelease(ctx, orgId, app.ID, releaseBranch, releaseVersion)
		if err != nil {
			return err
		}

		if !found {
			fmt.Println("Not found release for application", app.Name)
			continue
		}

		deployment, err := common.CreateRuntime(ctx, orgId, uint64(r.ProjectID), uint64(r.ApplicationID), workspace, r.ReleaseID)
		if err != nil {
			return err
		}

		rDeployments = append(rDeployments, deployment)
	}

	var taskRunners []dicedir.TaskRunner
	for _, deployment := range rDeployments {
		taskRunners = append(taskRunners, func() bool {
			status, err := common.GetDeploymentStatus(ctx, orgId, deployment.PipelineID)
			if err == nil && status == apistructs.PipelineStatusSuccess {
				return true
			}
			return false
		})
	}
	err = dicedir.DoTaskListWithTimeout(time.Duration(waitTime)*time.Minute, taskRunners)
	if err != nil {
		return err
	}

	return nil
}

type ReleaseApp struct {
	Branch  string `yaml:"branch"`
	Version string `yaml:"version"`
}
type ReleaseConfig struct {
	Applications map[string]ReleaseApp `yaml:"applications"`
}
