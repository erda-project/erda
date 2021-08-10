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

package pipelineymlv1

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1/pipelineymlvars"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1/steptasktype"
)

func TestTaskConfig_decodeAggregateStepTasks(t *testing.T) {
	y := New(getBytePipelineYml())
	err := y.Parse()
	require.NoError(t, err)
	aggTC := y.obj.Stages[0].TaskConfigs[0]
	require.Equal(t, len(aggTC), 1)
	steps, err := aggTC.decodeAggregateStepTasks()
	require.NoError(t, err)
	require.Equal(t, len(steps), 2)
	get, ok := steps[0].(*GetTask)
	require.True(t, ok)
	require.Equal(t, get.Get, "repo")
	_, ok = steps[1].(*CustomTask)
	require.True(t, ok)
}

func TestTaskConfig_decodeStepTask(t *testing.T) {
	getTC := TaskConfig{
		FieldGet.String():     "repo",
		FieldDisable.String(): false,
		FieldEnvs.String(): map[string]string{
			"e1": "v1",
		},
	}
	getStep, err := getTC.decodeSingleStepTask(steptasktype.GET)
	require.NoError(t, err)
	step, ok := getStep.(StepTask)
	require.True(t, ok)
	require.False(t, step.IsDisable())
	require.Equal(t, len(step.GetEnvs()), 1)
	require.Equal(t, step.GetEnvs()["e1"], "v1")
	require.Equal(t, step.Name(), "repo")
	_, ok = getStep.(*PutTask)
	require.False(t, ok)
	get, ok := getStep.(*GetTask)
	require.True(t, ok)
	require.False(t, *get.Disable)
	require.Equal(t, len(get.Envs), 1)
	require.Equal(t, get.Envs["e1"], "v1")
	require.Equal(t, get.Get, "repo")
}

func TestTaskConfig_decodeStepTaskWithValidate(t *testing.T) {
	y := New(getBytePipelineYml())
	y.SetPipelineID(123)
	err := y.Parse(WithBranch("develop"))
	require.NoError(t, err)
	aggTC := y.obj.Stages[0].TaskConfigs[0]
	steps, err := aggTC.decodeAggregateStepTasks()
	require.NoError(t, err)
	getTC := steps[0].GetTaskConfig()
	getStep, err := getTC.decodeStepTaskWithValidate(steptasktype.GET, y)
	require.NoError(t, err)
	get, ok := getStep.(*GetTask)
	require.True(t, ok)
	require.Equal(t, get.Branch, "develop")
}

func TestConvertObjectToTaskConfig(t *testing.T) {
	tc, err := convertObjectToTaskConfig(GetTask{Get: "repo"})
	require.NoError(t, err)
	require.Equal(t, tc[FieldGet.String()], "repo")
}
