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
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/dicedir"
)

// BUILD command
var PIPELINERUN = command.Command{
	Name:       "run",
	ParentName: "PIPELINE",
	ShortHelp:  "Create an pipeline and run it",
	Example:    `$ erda-cli pipeline run`,
	Flags: []command.Flag{
		command.StringFlag{Short: "b", Name: "branch",
			Doc:          "branch to create pipeline, default is current branch",
			DefaultValue: ""},
		command.StringFlag{Short: "f", Name: "filename",
			Doc:          "filename for 'pipeline.yml'",
			DefaultValue: path.Join(dicedir.ProjectPipelineDir, "pipeline.yml")},
	},
	Run: PipelineRun,
}

// Create an pipeline and run it
func PipelineRun(ctx *command.Context, branch, filename string) error {
	// 1. check if .git dir exists in current directory
	// 2. parse current branch
	// 3. create pipeline, run it
	if _, err := os.Stat(".git"); err != nil {
		return err
	}

	if branch == "" {
		b, err := common.GetWorkspaceBranch()
		if err != nil {
			return err
		}
		branch = b
	}

	// fetch appID
	orgName, projectName, appName, err := common.GetWorkspaceInfo(command.Remote)
	if err != nil {
		return err
	}

	org, err := common.GetOrgDetail(ctx, orgName)
	if err != nil {
		return err
	}

	orgID := strconv.FormatUint(org.Data.ID, 10)
	repoStats, err := common.GetRepoStats(ctx, orgID, projectName, appName)
	if err != nil {
		return err
	}

	var (
		request      apistructs.PipelineCreateRequest
		pipelineResp apistructs.PipelineCreateResponse
	)
	request.AppID = uint64(repoStats.Data.ApplicationID)
	request.Branch = branch
	request.Source = apistructs.PipelineSourceDice
	request.PipelineYmlSource = apistructs.PipelineYmlSourceGittar
	request.PipelineYmlName = filename
	request.AutoRun = true

	// create pipeline
	response, err := ctx.Post().Path("/api/cicds").JSONBody(request).Do().JSON(&pipelineResp)
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return errors.Errorf("build fail, status code: %d, err: %+v", response.StatusCode(), pipelineResp.Error)
	}
	if !pipelineResp.Success {
		return errors.Errorf("build fail: %+v", pipelineResp.Error)
	}
	ctx.Succ("run pipeline: %s for branch: %s, pipelineID: %d, you can view building status via `erda-cli status -i %d`",
		filename, branch, pipelineResp.Data.ID, pipelineResp.Data.ID)

	return nil
}
