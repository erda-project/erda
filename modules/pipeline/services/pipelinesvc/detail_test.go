// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package pipelinesvc

import (
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func Test_SimplePipelineBaseDetail(t *testing.T) {
	var tables = []struct {
		find  bool
		error bool
	}{
		{
			true,
			true,
		},
		{
			false,
			true,
		},
		{
			true,
			false,
		},
		{
			false,
			false,
		},
	}

	for _, data := range tables {
		var client = &dbclient.Client{}
		guard := monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetPipelineBase", func(client *dbclient.Client, id uint64, ops ...dbclient.SessionOption) (spec.PipelineBase, bool, error) {
			if data.error {
				return spec.PipelineBase{}, data.find, fmt.Errorf("")
			}
			return spec.PipelineBase{}, data.find, fmt.Errorf("")
		})
		p := PipelineSvc{dbClient: client}
		_, err := p.SimplePipelineBaseDetail(uint64(0))
		if data.find || data.error {
			assert.Error(t, err)
		}
		guard.Unpatch()
	}
}

func TestCanCancel(t *testing.T) {
	require.False(t, canCancel(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusInitializing}}))
	require.False(t, canCancel(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusAnalyzeFailed}}))
	require.False(t, canCancel(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusAnalyzed}}))
	require.True(t, canCancel(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusBorn}}))
	require.True(t, canCancel(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusQueue}}))
	require.True(t, canCancel(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusRunning}}))
	require.False(t, canCancel(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusTimeout}}))
}

func TestCanForceCancel(t *testing.T) {
	require.False(t, canForceCancel(spec.Pipeline{}))
}

func TestCanRerun(t *testing.T) {
	require.False(t, canRerun(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusInitializing}}))
	require.False(t, canRerun(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusAnalyzed}}))
	require.True(t, canRerun(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusAnalyzeFailed}}))
	require.False(t, canRerun(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusQueue}}))
	require.False(t, canRerun(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusRunning}}))
	require.True(t, canRerun(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusSuccess}}))
	require.True(t, canRerun(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusFailed}}))
}

func TestCanRerunFailed(t *testing.T) {
	require.False(t, canRerunFailed(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusInitializing}}))
	require.False(t, canRerunFailed(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusAnalyzed}}))
	require.True(t, canRerunFailed(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusAnalyzeFailed}}))
	require.False(t, canRerunFailed(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusQueue}}))
	require.False(t, canRerunFailed(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusRunning}}))
	require.False(t, canRerunFailed(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusSuccess}}))
	require.True(t, canRerunFailed(spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusFailed}}))
}

// func TestCanStartCron(t *testing.T) {
// 	require.False(t, canStartCron(spec.Pipeline{Type: spec.PipelineTypeNormal}, nil))
// 	require.True(t, canStartCron(spec.Pipeline{Type: spec.PipelineTypeCron}, &spec.PipelineCron{Enable: &[]bool{false}[0]}))
// 	require.False(t, canStartCron(spec.Pipeline{Type: spec.PipelineTypeCron}, &spec.PipelineCron{Enable: &[]bool{true}[0]}))
// }
//
// func TestCanStopCron(t *testing.T) {
// 	require.False(t, canStopCron(spec.Pipeline{Type: spec.PipelineTypeNormal}, nil))
// 	require.False(t, canStopCron(spec.Pipeline{Type: spec.PipelineTypeCron}, &spec.PipelineCron{Enable: &[]bool{false}[0]}))
// 	require.True(t, canStopCron(spec.Pipeline{Type: spec.PipelineTypeCron}, &spec.PipelineCron{Enable: &[]bool{true}[0]}))
// }

func TestCanPause(t *testing.T) {
	require.False(t, canPause(spec.Pipeline{}))
}

func TestCanUnpause(t *testing.T) {
	require.False(t, canUnpause(spec.Pipeline{}))
}

func TestFindRunningStageID(t *testing.T) {
	p := spec.Pipeline{PipelineBase: spec.PipelineBase{Status: apistructs.PipelineStatusRunning}}

	// 1 R
	// 2 S => 3
	// 3 S
	id1 := findRunningStageID(p, []spec.PipelineTask{
		{StageID: 1, Status: apistructs.PipelineStatusRunning},
		{StageID: 2, Status: apistructs.PipelineStatusSuccess},
		{StageID: 3, Status: apistructs.PipelineStatusSuccess},
	})
	assert.True(t, id1 == 3)

	// 1 R
	// 2 A => 1
	// 3 A
	id2 := findRunningStageID(p, []spec.PipelineTask{
		{StageID: 1, Status: apistructs.PipelineStatusRunning},
		{StageID: 2, Status: apistructs.PipelineStatusAnalyzed},
		{StageID: 3, Status: apistructs.PipelineStatusAnalyzed},
	})
	assert.True(t, id2 == 1)
}

func TestIsEventsContainWarn(t *testing.T) {
	normalEvents := "Events:\n Type    Reason     Age   From               Message\n ----    ------     ----  ----               -------\n Normal  Scheduled  7s    default-scheduler  Successfully assigned pipeline-4152/pipeline-4152.pipeline-task-8296-tgxd7 to node-010000006200\n Normal  Pulled     6s    kubelet            Container image \"registry.erda.cloud/erda-actions/action-agent:1.2-20210804-75232495\" already present on machine"
	warnEvents := "Events:\n Type    Reason     Age   From               Message\n ----    ------     ----  ----               -------\n Warning  Scheduled  7s    default-scheduler  Successfully assigned pipeline-4152/pipeline-4152.pipeline-task-8296-tgxd7 to node-010000006200\n Normal  Pulled     6s    kubelet            Container image \"registry.erda.cloud/erda-actions/action-agent:1.2-20210804-75232495\" already present on machine"
	shortEvents := "Events:\nType    Reason     Age   From               Message\n----    ------     ----  ----               -------"
	assert.Equal(t, false, isEventsContainWarn(normalEvents))
	assert.Equal(t, true, isEventsContainWarn(warnEvents))
	assert.Equal(t, false, isEventsContainWarn(shortEvents))
}
