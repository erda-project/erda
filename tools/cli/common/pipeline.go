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

package common

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/terminal/color_str"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

func GetPipeline(ctx *command.Context, pipelineID uint64) (pipelinepb.PipelineDetailDTO, error) {
	// fetch pipeline info
	var pipelineInfoResp pipelinepb.PipelineDetailResponse
	response, err := ctx.Get().
		Path(fmt.Sprintf("/api/pipelines/%d", pipelineID)).
		Do().JSON(&pipelineInfoResp)
	if err != nil {
		return pipelinepb.PipelineDetailDTO{}, err
	}
	if !response.IsOK() {
		return pipelinepb.PipelineDetailDTO{}, errors.Errorf("status fail, status code: %d", response.StatusCode())
	}

	return *pipelineInfoResp.Data, nil
}

// BuildCheckLoop checks build status in a loop while interactive is true
func BuildCheckLoop(ctx *command.Context, pipelineID uint64) error {
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
		pipelineInfo, err := GetPipeline(ctx, pipelineID)
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

				switch apistructs.PipelineStatus(task.Status) {
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

		pipelineStatus := apistructs.PipelineStatus(pipelineInfo.Status)
		if pipelineStatus == apistructs.PipelineStatusStopByUser ||
			pipelineStatus == apistructs.PipelineStatusFailed ||
			pipelineStatus == apistructs.PipelineStatusTimeout {
			fmt.Print(color_str.Green(fmt.Sprintf("build faild, status: %s, time elapsed: %s\n",
				pipelineInfo.Status, utils.ToTimeSpanString(int(pipelineInfo.CostTimeSec)))))
			var msg = "nil"
			if showMessage := pipelineInfo.Extra.ShowMessage; showMessage != nil {
				fmt.Println(showMessage.Stacks)
				msg = showMessage.Msg
			}
			return fmt.Errorf(utils.FormatErrMsg("pipeline info",
				"build error: "+msg, false))
		}

		if pipelineStatus == apistructs.PipelineStatusSuccess {
			fmt.Print(color_str.Green(fmt.Sprintf("\nbuild succ, time elapsed: %s\n", utils.ToTimeSpanString(int(pipelineInfo.CostTimeSec)))))
			return nil
		}

		time.Sleep(time.Second * 2)
	}
}
