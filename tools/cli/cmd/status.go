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
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var STATUS = command.Command{
	Name:      "status",
	ShortHelp: "Show build status",
	Example: `
  $ erda-cli status -b develop
`,
	Flags: []command.Flag{
		command.StringFlag{Short: "b", Name: "branch", Doc: "specify branch to show pipeline status, default is current branch", DefaultValue: ""},
		command.IntFlag{Short: "i", Name: "pipelineID", Doc: "specify pipeline id to show pipeline status", DefaultValue: 0},
	},
	Run: RunPipelineStatus,
}

// RunBuildsInspect displays detailed information on the build record
func RunPipelineStatus(ctx *command.Context, branch string, pipelineID int) error {
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

	// TODO gittar-adaptor 提供 API 根据 branch & git remote url 查询 pipelineID
	// fetch appID
	orgName, projectName, appName, err := common.GetWorkspaceInfo()
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
	// fetch ymlName path
	var pipelineCombResp apistructs.PipelineInvokedComboResponse
	response, err := ctx.Get().Path("/api/cicds/actions/app-invoked-combos").
		Param("appID", strconv.FormatInt(repoStats.Data.ApplicationID, 10)).
		Do().JSON(&pipelineCombResp)
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return errors.Errorf("status fail, status code: %d, err: %+v", response.StatusCode(), pipelineCombResp.Error)
	}
	if !pipelineCombResp.Success {
		return errors.Errorf("status fail: %+v", pipelineCombResp.Error)
	}
	var ymlName string
	for _, v := range pipelineCombResp.Data {
		if v.Branch == branch {
			for _, item := range v.PagingYmlNames {
				if item != apistructs.DefaultPipelineYmlName {
					ymlName = item
				}
			}
		}
	}

	// fetch pipelineID
	var pipelineListResp apistructs.PipelinePageListResponse
	response, err = ctx.Get().Path("/api/cicds").
		Param("appID", strconv.FormatInt(repoStats.Data.ApplicationID, 10)).
		Param("sources", "dice").
		Param("ymlNames", ymlName).
		Param("branches", branch).
		Param("pageNo", "1").
		Param("pageSize", "1").
		Do().JSON(&pipelineListResp)
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return errors.Errorf("status fail, status code: %d, err: %+v", response.StatusCode(), pipelineListResp.Error)
	}
	if !pipelineListResp.Success {
		return errors.Errorf("status fail: %+v", pipelineListResp.Error)
	}

	if len(pipelineListResp.Data.Pipelines) == 0 {
		return errors.Errorf("status fail, can't find pipeline info")
	}

	if pipelineID == 0 {
		pipelineID = int(pipelineListResp.Data.Pipelines[0].ID)
	}

	// fetch pipeline info
	var pipelineInfoResp apistructs.PipelineDetailResponse
	response, err = ctx.Get().Path(fmt.Sprintf("/api/pipelines/%d", pipelineID)).
		Do().JSON(&pipelineInfoResp)
	if err != nil {
		return err
	}
	if !response.IsOK() {
		return errors.Errorf("status fail, status code: %d, err: %+v", response.StatusCode(), pipelineInfoResp.Error)
	}
	if !pipelineInfoResp.Success {
		return errors.Errorf("status fail: %+v", pipelineInfoResp.Error)
	}

	data := [][]string{}
	var currentStageIndex int
	for i, stage := range pipelineInfoResp.Data.PipelineStages {
		currentStageIndex = i
		stageDone := true
		for _, task := range stage.PipelineTasks {
			if task.Status != apistructs.PipelineStatusSuccess {
				stageDone = false
				break
			}
		}

		if !stageDone {
			for _, task := range stage.PipelineTasks {
				data = append(data, []string{
					strconv.Itoa(pipelineID),
					strconv.FormatUint(task.ID, 10),
					task.Name,
					task.Status.String(),
					task.TimeBegin.Format("2006-01-02 15:04:05"),
				})
			}
			break
		}
	}

	fmt.Printf("pipeline progress (currentStage/totalStages): %d/%d\n\n",
		currentStageIndex+1, len(pipelineInfoResp.Data.PipelineStages))

	return table.NewTable().Header([]string{
		"pipelineID", "taskID", "taskName", "taskStatus", "startedAt",
	}).Data(data).Flush()
}
