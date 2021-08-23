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

package statusutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestCalculatePipelineTaskStatus(t *testing.T) {
	t.Log("init an empty pipeline task without spec.Status")
	task := &spec.PipelineTask{}
	require.Equal(t, apistructs.PipelineEmptyStatus, task.Status)
	_, err := CalculatePipelineTaskStatus(task)
	require.Error(t, err)

	t.Log("set spec.Status=failed")
	task.Status = apistructs.PipelineStatusFailed
	s, err := CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusFailed, s)

	t.Log("set spec.Status=error and allowFailure=true")
	// 设置 spec.Status = failed 并且
	task.Status = apistructs.PipelineStatusError
	task.Extra.AllowFailure = true
	s, err = CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("set spec.Status=unknown and allowFailure=false")
	task.Status = apistructs.PipelineStatusUnknown
	task.Extra.AllowFailure = false
	s, err = CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusUnknown, s)

	t.Log("set spec.Status=disabled")
	task.Status = apistructs.PipelineStatusDisabled
	s, err = CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusDisabled, s)

	t.Log("========test not end spec.Status========")

	t.Log("set spec.Status=born")
	task.Status = apistructs.PipelineStatusBorn
	s, err = CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusBorn, s)

	t.Log("set allowFailure=true")
	task.Extra.AllowFailure = true
	s, err = CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusBorn, s)

	t.Log("set spec.Status=running")
	task.Status = apistructs.PipelineStatusRunning
	s, err = CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set spec.Status=dbError")
	task.Status = apistructs.PipelineStatusDBError
	s, err = CalculatePipelineTaskStatus(task)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)
}

func TestCalculatePipelineStageStatus(t *testing.T) {
	t.Log("init an empty stage without spec.Status and tasks")
	stage := &spec.PipelineStageWithTask{}
	s, err := CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("set stage spec.Status=failed, but it should have no effect when calculate stage spec.Status")
	stage.Status = apistructs.PipelineStatusFailed
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("add task without spec.Status")
	stage.PipelineTasks = append(stage.PipelineTasks, &spec.PipelineTask{})
	_, err = CalculatePipelineStageStatus(stage)
	require.Error(t, err)

	t.Log("remove no spec.Status task")
	stage.PipelineTasks = stage.PipelineTasks[1:]
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("add task[0]: spec.Status=born")
	stage.PipelineTasks = append(stage.PipelineTasks, &spec.PipelineTask{Status: apistructs.PipelineStatusBorn})
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusBorn, s)

	t.Log("add task[1]: spec.Status=born")
	stage.PipelineTasks = append(stage.PipelineTasks, &spec.PipelineTask{Status: apistructs.PipelineStatusBorn})
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusBorn, s)

	t.Log("add task[2]: spec.Status=queue")
	stage.PipelineTasks = append(stage.PipelineTasks, &spec.PipelineTask{Status: apistructs.PipelineStatusQueue})
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set task[0]=running, task[1]=failed, task[2]=success")
	stage.PipelineTasks[0].Status = apistructs.PipelineStatusRunning
	stage.PipelineTasks[1].Status = apistructs.PipelineStatusFailed
	stage.PipelineTasks[2].Status = apistructs.PipelineStatusSuccess
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set task[0]=timeout")
	stage.PipelineTasks[0].Status = apistructs.PipelineStatusTimeout
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusFailed, s)

	t.Log("set task[0]=success, task[1]=success")
	stage.PipelineTasks[0].Status = apistructs.PipelineStatusSuccess
	stage.PipelineTasks[1].Status = apistructs.PipelineStatusSuccess
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("add task[3]: spec.Status=born")
	stage.PipelineTasks = append(stage.PipelineTasks, &spec.PipelineTask{Status: apistructs.PipelineStatusBorn})
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusBorn, s)

	t.Log("set task[3]: spec.Status=paused")
	stage.PipelineTasks[3].Status = apistructs.PipelineStatusPaused
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusPaused, s)

	t.Log("set task[3]: spec.Status=running")
	stage.PipelineTasks[3].Status = apistructs.PipelineStatusRunning
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set task[3]: spec.Status=failed")
	stage.PipelineTasks[3].Status = apistructs.PipelineStatusFailed
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusFailed, s)

	t.Log("set task[3]: allowFailure=true")
	stage.PipelineTasks[3].Extra.AllowFailure = true
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("========test another stage========")
	stage = &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusPaused}, {Status: apistructs.PipelineStatusSuccess}}}
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusPaused, s)

	stage.PipelineTasks[0].Status = apistructs.PipelineStatusBorn
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusBorn, s)

	t.Log("========test another stage with disable========")
	stage = &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusPaused}, {Status: apistructs.PipelineStatusDisabled}, {Status: apistructs.PipelineStatusBorn}}}
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("remove task[2]: spec.Status=born")
	stage.PipelineTasks = stage.PipelineTasks[:2]
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusPaused, s)

	t.Log("remove task[0]: spec.Status=paused")
	stage.PipelineTasks = stage.PipelineTasks[1:]
	s, err = CalculatePipelineStageStatus(stage)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)
}

func TestCalculateStatus(t *testing.T) {
	t.Log("init an empty p without spec.Status and stages")
	p := spec.PipelineWithStage{}
	s, err := CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("set p spec.Status=failed but it should have no effect when calculate p spec.Status")
	p.Status = apistructs.PipelineStatusFailed
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("add stage[0]: task[0]=born")
	p.PipelineStages = append(p.PipelineStages, &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusBorn}}})
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set stage[0]: task[0]=born, task[1]=failed and allowFailure=true")
	p.PipelineStages[0].PipelineTasks = []*spec.PipelineTask{{Status: apistructs.PipelineStatusBorn}, {Status: apistructs.PipelineStatusFailed, Extra: spec.PipelineTaskExtra{AllowFailure: true}}}
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set stage[0]: task[0]=failed")
	p.PipelineStages[0].PipelineTasks[0].Status = apistructs.PipelineStatusFailed
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusFailed, s)

	t.Log("set stage[0]: task[0]=disabled")
	p.PipelineStages[0].PipelineTasks[0].Status = apistructs.PipelineStatusDisabled
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("set stage[0]: task[0]=success")
	p.PipelineStages[0].PipelineTasks[0].Status = apistructs.PipelineStatusSuccess
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusSuccess, s)

	t.Log("add stage[1]: task[0]=paused")
	p.PipelineStages = append(p.PipelineStages, &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusPaused}}})
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusPaused, s)

	t.Log("add stage[2]: task[0]=born, task[1]=born")
	p.PipelineStages = append(p.PipelineStages, &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusBorn}, {Status: apistructs.PipelineStatusBorn}}})
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusPaused, s)

	t.Log("unpause stage[1]")
	p.PipelineStages[1].PipelineTasks[0].Status = apistructs.PipelineStatusRunning
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set stage[1]: task[0]=success")
	p.PipelineStages[1].PipelineTasks[0].Status = apistructs.PipelineStatusSuccess
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusRunning, s)

	t.Log("set stage[2]: task[0]=success, task[1]=timeout")
	p.PipelineStages[2].PipelineTasks[0].Status = apistructs.PipelineStatusSuccess
	p.PipelineStages[2].PipelineTasks[1].Status = apistructs.PipelineStatusTimeout
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusFailed, s)

	t.Log("add stage[3]: task[0]=disabled")
	p.PipelineStages = append(p.PipelineStages, &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusDisabled}}})
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusFailed, s)

	t.Log("add stage[4]: task[0]=noNeedBySystem")
	t.Log("add stage[5]: task[0]=noNeedBySystem")
	p.PipelineStages = append(p.PipelineStages, &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusNoNeedBySystem}}})
	p.PipelineStages = append(p.PipelineStages, &spec.PipelineStageWithTask{PipelineTasks: []*spec.PipelineTask{{Status: apistructs.PipelineStatusNoNeedBySystem}}})
	fmt.Println(len(p.PipelineStages))
	s, err = CalculateStatus(p)
	require.NoError(t, err)
	require.Equal(t, apistructs.PipelineStatusFailed, s)
}
