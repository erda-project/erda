package pipelinesvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

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
