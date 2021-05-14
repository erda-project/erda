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

package dbclient

//import (
//	"fmt"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestClient_UpdatePipelineTask(t *testing.T) {
//	task, err := client.GetPipelineTask(1)
//	require.NoError(t, err)
//	fmt.Println(task.CostTimeSec)
//	fmt.Println(task.Status)
//
//	task.CostTimeSec = 100
//	task.QueueTimeSec = 20
//	task.Extra.Namespace = "xxx-xxx"
//	task.Status = apistructs.PipelineStatusFailed
//	err = client.UpdatePipelineTask(task.ID, &task)
//	require.NoError(t, err)
//	require.True(t, task.Status == apistructs.PipelineStatusFailed)
//	require.True(t, task.CostTimeSec == 100)
//	require.True(t, task.QueueTimeSec == 20)
//	require.True(t, task.Extra.Namespace == "xxx-xxx")
//}
