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
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/table"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

var VIEW = command.Command{
	Name:      "view",
	ShortHelp: "View build status",
	Example: `
  $ erda-cli view -b develop -i <pipelineID> --host https://openapi.erda.cloud
`,
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "r",
			Name:         "repo",
			Doc:          "the repo on Erda DOP, default current repo.",
			DefaultValue: os.Getenv("ERDA_REPO"),
		},
		command.StringFlag{Short: "b", Name: "branch", Doc: "specify branch to show pipeline status, default is current branch", DefaultValue: ""},
		command.IntFlag{Short: "i", Name: "pipelineID", Doc: "specify pipeline id to show pipeline status", DefaultValue: 0},
		command.BoolFlag{
			Short:        "w",
			Name:         "watch",
			Doc:          "watch the status",
			DefaultValue: false,
		},
	},
	Run: RunViewPipe,
}

// RunViewPipe displays detailed information on the build record
func RunViewPipe(ctx *command.Context, repo, branch string, pipelineID int, watch bool) (err error) {
	if watch {
		_ = BuildCheckLoop(ctx, strconv.FormatInt(int64(pipelineID), 10))
	}

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
	var orgName, projectName, appName string
	if repo != "" {
		orgName, projectName, appName, err = common.GetWorkspaceInfoFromErdaRepo(repo)
	} else {
		orgName, projectName, appName, err = common.GetWorkspaceInfo()
	}
	if err != nil {
		return errors.Wrapf(err, "failed to GetWorksapceInfo, repo: %s", repo)
	}

	org, err := common.GetOrgDetail(ctx, orgName)
	if err != nil {
		return err
	}

	orgID := strconv.FormatUint(org.Data.ID, 10)
	repoStats, err := common.GetRepoStats(ctx, orgID, projectName, appName)
	if err != nil {
		return errors.Wrapf(err, "orgID: %v, projectName: %s, appName: %s", orgID, projectName, appName)
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

	var (
		data              [][]string
		currentStageIndex int
		total             = len(pipelineInfoResp.Data.PipelineStages)
	)
	for i, stage := range pipelineInfoResp.Data.PipelineStages {
		success := true
		for _, task := range stage.PipelineTasks {
			success = success && task.Status.IsSuccessStatus()
			data = append(data, []string{
				strconv.Itoa(pipelineID),
				strconv.FormatUint(task.ID, 10),
				task.Name,
				task.Status.String(),
				task.TimeBegin.Format("2006-01-02 15:04:05"),
			})
		}
		if success {
			currentStageIndex = i
		}
	}
	fmt.Printf("Pipeline progress (current/total): %d/%d\n",
		currentStageIndex+1, total)
	if err = table.NewTable().Header([]string{
		"pipelineID", "taskID", "taskName", "taskStatus", "startedAt",
	}).Data(data).Flush(); err != nil {
		return err
	}
	printMetadata(pipelineInfoResp.Data.PipelineStages)
	seeMore(ctx, orgName, int(repoStats.Data.ProjectID), int(repoStats.Data.ApplicationID), branch, pipelineID, ymlName)
	return nil
}

func seeMore(ctx *command.Context, orgName string, projectID, appID int, branch string, pipelineID int, pipelineName string) {
	frontUrl := ctx.Get().Path(fmt.Sprintf("/%s/dop/projects/%v/apps/%v/pipeline",
		orgName, projectID, appID)).
		Param("pipelineID", strconv.Itoa(pipelineID)).
		Param("nodeId", base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{
			strconv.FormatInt(int64(projectID), 10),
			strconv.FormatInt(int64(appID), 10),
			"tree",
			branch,
			pipelineName,
		}, "/")))).GetUrl()
	frontUrl = strings.TrimPrefix(frontUrl, "https://")
	frontUrl = strings.TrimPrefix(frontUrl, "http://")
	if u, err := url.Parse(frontUrl); err == nil {
		u.Host = strings.ReplaceAll(u.Host, "openapi.", "")
		fmt.Printf("see more at %s\n", u.String())
	} else {
		fmt.Printf("failed to parse %s", frontUrl)
	}
}

func printMetadata(pipelineStages []apistructs.PipelineStageDetailDTO) {
	var meta []map[string]map[string]string
	for _, pipelineStage := range pipelineStages {
		s := make(map[string]map[string]string)
		for _, pipelineTask := range pipelineStage.PipelineTasks {
			t := make(map[string]string)
			for _, metadata := range pipelineTask.Result.Metadata {
				t[metadata.Name] = metadata.Value
			}
			s[pipelineTask.Name] = t
		}
		meta = append(meta, s)
	}
	fmt.Println("\nPipeline Metadata")
	_ = yaml.NewEncoder(os.Stdout).Encode(meta)
}
