package dbclient

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/apistructs"
)

func TestClient_UpdatePipelineTask(t *testing.T) {
	task, err := client.GetPipelineTask(1)
	require.NoError(t, err)
	fmt.Println(task.CostTimeSec)
	fmt.Println(task.Status)

	task.CostTimeSec = 100
	task.QueueTimeSec = 20
	task.Extra.Namespace = "xxx-xxx"
	task.Status = apistructs.PipelineStatusFailed
	err = client.UpdatePipelineTask(task.ID, &task)
	require.NoError(t, err)
	require.True(t, task.Status == apistructs.PipelineStatusFailed)
	require.True(t, task.CostTimeSec == 100)
	require.True(t, task.QueueTimeSec == 20)
	require.True(t, task.Extra.Namespace == "xxx-xxx")
}
