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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/format"
)

// BUILD command
var BUILD = command.Command{
	Name:      "build",
	ShortHelp: "Create an pipeline and run it",
	Example:   `$ erda-cli build`,
	Flags: []command.Flag{
		command.StringFlag{
			Short:        "r",
			Name:         "repo",
			Doc:          "the repo on Erda DOP",
			DefaultValue: os.Getenv("ERDA_REPO"),
		},
		command.StringFlag{Short: "b", Name: "branch", Doc: "branch to create pipeline, default is current branch", DefaultValue: ""},
		command.StringFlag{
			Short:        "f",
			Name:         "pipeline-file",
			Doc:          "The specified local pipeline file",
			DefaultValue: "",
		},
		command.StringFlag{
			Short:        "",
			Name:         "alias",
			Doc:          "pipeline task alias name",
			DefaultValue: "",
		},
	},
	Run: RunBuild,
}

// RunBuild Create an pipeline and run it
func RunBuild(ctx *command.Context, repo, branch, filename, alias string) (err error) {
	// 1. check if .git dir exists in current directory
	// 2. parse current branch
	// 3. create pipeline, run it
	if _, err = os.Stat(".git"); err != nil {
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
	if repo == "" {
		fmt.Printf("repo not specified, use origin")
		repo = common.GetOriginRepo()
	}
	if repo == "" {
		return errors.New("repo not found. Exit.")
	}
	orgName, projectName, appName, err := common.GetWorkspaceInfoFromErdaRepo(repo)
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
	request.AutoRun = true

	if filename != "" {
		content, err := parsePipeline(filename)
		if err != nil {
			return errors.Wrapf(err, "failed to parsePipeline: %s", filename)
		}
		request.PipelineYmlSource = apistructs.PipelineYmlSourceContent
		request.PipelineYmlName = strings.TrimSuffix(path.Base(filename), path.Ext(filename))
		request.PipelineYmlContent = string(content)
	}
	if alias != "" {
		request.PipelineYmlName = alias
	}

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
	ctx.Succ("building for branch: %s, pipelineID: %d, you can view building status via \nerda-cli view -w -i %d -b %s -r %s --host %s",
		branch, pipelineResp.Data.ID, pipelineResp.Data.ID, branch, repo, ctx.CurrentOpenApiHost)

	return nil
}

// BuildCheckLoop checks build status in a loop while interactive is true
func BuildCheckLoop(ctx *command.Context, buildID string) error {
	type taskInfoHinted struct {
		Running bool
		Fail    bool
		Done    bool
	}

	var (
		stageDone        = map[uint64]struct{}{}
		taskDone         = map[uint64]struct{}{}
		taskInfoOutputed = map[uint64]taskInfoHinted{}
		currentStage     uint64
	)

	for {
		pipelineInfo, err := common.GetBuildDetail(ctx, buildID)
		if err != nil {
			return err
		}

		for i, stage := range pipelineInfo.PipelineStages {
			if _, ok := stageDone[stage.ID]; ok {
				continue
			}

			if stage.ID != currentStage {
				fmt.Print(color_str.Green(fmt.Sprintf("\u2739 Stage %d\n", i)))
				currentStage = stage.ID
			}

			currentStageDone := true
			for _, task := range stage.PipelineTasks {
				if _, ok := taskDone[task.ID]; ok {
					continue
				}

				v, ok := taskInfoOutputed[task.ID]
				if !ok {
					taskInfoOutputed[task.ID] = taskInfoHinted{}
					v = taskInfoHinted{}
				}

				switch task.Status {
				case apistructs.PipelineStatusQueue, apistructs.PipelineStatusRunning:
					currentStageDone = false
					if !v.Running {
						fmt.Print(color_str.Green(fmt.Sprintf("    \u2615 Run task: %s\n", task.Name)))
						v.Running = true
					}
				case apistructs.PipelineStatusStopByUser, apistructs.PipelineStatusFailed, apistructs.PipelineStatusTimeout:
					currentStageDone = false
					if !v.Fail {
						fmt.Print(color_str.Red(fmt.Sprintf("    \u2718 Fail task: %s\n", task.Name)))
						v.Fail = true
					}
				case apistructs.PipelineStatusSuccess:
					currentStageDone = true
					if !v.Done {
						fmt.Print(color_str.Green(fmt.Sprintf("    \u2714 Success task: %s\n", task.Name)))
						taskDone[task.ID] = struct{}{}
						v.Done = true
					}
				default:
					currentStageDone = false
				}
				taskInfoOutputed[task.ID] = v
			}

			if !currentStageDone { // current stage is not done, don't need check next stage
				break
			}
			stageDone[stage.ID] = struct{}{}
		}

		if pipelineInfo.Status == apistructs.PipelineStatusStopByUser ||
			pipelineInfo.Status == apistructs.PipelineStatusFailed ||
			pipelineInfo.Status == apistructs.PipelineStatusTimeout {
			fmt.Print(color_str.Green(fmt.Sprintf("build faild, status: %s, time elapsed: %s\n",
				pipelineInfo.Status, format.ToTimeSpanString(int(pipelineInfo.CostTimeSec)))))
			var msg = "nil"
			if showMessage := pipelineInfo.Extra.ShowMessage; showMessage != nil {
				fmt.Println(showMessage.Stacks)
				msg = showMessage.Msg
			}
			return fmt.Errorf(format.FormatErrMsg("pipeline info",
				"build error: "+msg, false))
		}

		if pipelineInfo.Status == apistructs.PipelineStatusSuccess {
			fmt.Print(color_str.Green(fmt.Sprintf("\nbuild succ, time elapsed: %s\n", format.ToTimeSpanString(int(pipelineInfo.CostTimeSec)))))
			return nil
		}

		time.Sleep(time.Second * 2)
	}
}

// PrintWorkFlow prints build work flow in diagram format
func PrintWorkFlow(ctx *command.Context, buildID string) error {
	preLen := 0
	maxTaskNum := 0
	maxTaskLen := 0

	pipelineInfo, err := common.GetBuildDetail(ctx, buildID)
	if err != nil {
		fmt.Print(color_str.Red("获取构建信息失败，请手动查询公构建状态\n"))
		return err
	}

	for i := range pipelineInfo.PipelineStages {
		if preLen < (len(pipelineInfo.PipelineStages[i].Name) + 2) {
			preLen = len(pipelineInfo.PipelineStages[i].Name) + 2
		}

		if maxTaskLen < len(pipelineInfo.PipelineStages[i].PipelineTasks) {
			maxTaskNum = len(pipelineInfo.PipelineStages[i].PipelineTasks)
		}
		for j := range pipelineInfo.PipelineStages[i].PipelineTasks {
			if maxTaskLen < (len(pipelineInfo.PipelineStages[i].PipelineTasks[j].Name) + 2) {
				maxTaskLen = len(pipelineInfo.PipelineStages[i].PipelineTasks[j].Name) + 2
			}
		}
	}

	for stageID := range pipelineInfo.PipelineStages {
		if stageID != 0 {
			for i := 0; i < 2; i++ {
				for h := 0; h < (preLen + (maxTaskNum*(maxTaskLen+5))/2 - 3 + maxTaskLen/2 + 5); h++ {
					fmt.Print(" ")
				}
				fmt.Print(color_str.Green("|"))
				fmt.Print(color_str.Green("|"))
				fmt.Print("\n")
			}

			for h := 0; h < (preLen + (maxTaskNum*(maxTaskLen+5))/2 - 3 + maxTaskLen/2 + 5); h++ {
				fmt.Print(" ")
			}
			fmt.Print(color_str.Green("\\"))
			fmt.Print(color_str.Green("/"))
			fmt.Print("\n")
		}

		var task []string
		stage := pipelineInfo.PipelineStages[stageID].Name

		for taskID := range pipelineInfo.PipelineStages[stageID].PipelineTasks {
			task = append(task, pipelineInfo.PipelineStages[stageID].PipelineTasks[taskID].Name)
		}

		if len(task) < 1 {
			return fmt.Errorf(format.FormatErrMsg("builds", "no tasks found of stage "+stage, false))
		}

		taskNum := len(task)
		taskPrintPoint := maxTaskNum * (maxTaskLen + 5) / taskNum

		for k := 0; k < 5; k++ {
			if k == 0 {
				fmt.Print(color_str.White(stage))

				blankLen := preLen - len(stage)
				for i := 0; i < (blankLen - 1); i++ {
					fmt.Print(" ")
				}

			} else {
				for i := 0; i < (preLen); i++ {
					fmt.Print(" ")
				}
			}

			for i := 0; i < 5; i++ {
				fmt.Print(" ")
			}

			for j := 0; j < taskNum; j++ {
				switch {
				case k == 2:
					for m := 0; m < ((taskPrintPoint - len(task[j])) / 2); m++ {
						fmt.Print(" ")
					}

					fmt.Print(color_str.Green("|"))
					fmt.Print(" ")
					fmt.Print(color_str.White(task[j]))
					fmt.Print(" ")
					fmt.Print(color_str.Green("|"))

					for i := 0; i < 5; i++ {
						fmt.Print(" ")
					}
				case (k % 2) == 1:
					for m := 0; m < ((taskPrintPoint - len(task[j])) / 2); m++ {
						fmt.Print(" ")
					}

					fmt.Print(color_str.Green("|"))
					fmt.Print(" ")
					for i := 0; i < len(task[j]); i++ {
						fmt.Print(" ")
					}
					fmt.Print(" ")
					fmt.Print(color_str.Green("|"))

					for i := 0; i < 5; i++ {
						fmt.Print(" ")
					}
				default:
					for m := 0; m < ((taskPrintPoint - len(task[j])) / 2); m++ {
						fmt.Print(" ")
					}

					fmt.Print(" ")

					for i := 0; i < (len(task[j]) + 2); i++ {
						fmt.Print(color_str.Green("-"))
					}
					fmt.Print(" ")

					for i := 0; i < 5; i++ {
						fmt.Print(" ")
					}
				}
			}
			fmt.Print("\n")
		}
	}

	return nil
}

func parsePipeline(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", errors.Wrapf(err, "failed to ReadFile: %s", filename)
	}

	branch, _ := common.GetWorkspaceBranch()
	content = bytes.ReplaceAll(content, []byte("((branch))"), []byte(branch))

	remotes := regexp.MustCompile(`\(\(remote\..*\)\)`).FindAllString(string(content), -1)
	for _, remote := range remotes {
		ss := strings.Split(remote, ".")
		if len(ss) < 2 {
			continue
		}
		remoteName := strings.TrimRight(ss[1], ")")
		out, err := exec.Command("git", "config", "--get", "remote."+remoteName+".url").CombinedOutput()
		if err != nil {
			return "", errors.Wrapf(err, "failed to get git remote url, remote name: %s", remoteName)
		}
		content = bytes.ReplaceAll(content, []byte("((remote."+remoteName+"))"), out)
	}

	return string(content), nil
}
