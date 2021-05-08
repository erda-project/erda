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

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"
	"github.com/xormplus/xorm"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

type result struct {
	InsertId int64
}

func (this result) LastInsertId() (int64, error) {
	return this.InsertId, nil
}

// RowsAffected returns the number of rows affected by an
// update, insert, or delete. Not every database or database
// driver may support this.
func (this result) RowsAffected() (int64, error) {
	return 0, nil
}

func TestBatchCreatePipelineTask(t *testing.T) {
	var tables = []struct {
		num      int
		resultID int64
	}{
		{
			1000,
			102001,
		},
		{
			304,
			11001,
		},
		{
			40,
			12001,
		},
		{
			1,
			200021,
		},
	}
	for _, data := range tables {

		var client Client
		engine := xorm.Engine{}
		client.Engine = &engine
		var session = xorm.Session{}

		var tasks []*spec.PipelineTask
		for i := 0; i < data.num; i++ {
			tasks = append(tasks, &spec.PipelineTask{})
		}

		// closure to mock sql result lastInsertId
		var lastInsertId = 0
		var sqlRunTimes = 0
		var constructLastInsertId = func() result {
			var result result
			if lastInsertId == 0 {
				lastInsertId = int(data.resultID)
			} else {
				lastInsertId += BatchInsertTaskNum
			}
			result.InsertId = int64(lastInsertId)
			sqlRunTimes++
			return result
		}
		patch5 := monkey.PatchInstanceMethod(reflect.TypeOf(&session), "Close", func(s *xorm.Session) {
			return
		})

		patch4 := monkey.PatchInstanceMethod(reflect.TypeOf(&session), "Commit", func(s *xorm.Session) error {
			return nil
		})

		patch3 := monkey.PatchInstanceMethod(reflect.TypeOf(&session), "Begin", func(s *xorm.Session) error {
			return nil
		})

		patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(&session), "Rollback", func(s *xorm.Session) error {
			return nil
		})

		patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(&session), "Exec", func(s *xorm.Session, args ...interface{}) (sql.Result, error) {
			return constructLastInsertId(), nil
		})

		patch := monkey.PatchInstanceMethod(reflect.TypeOf(&engine), "NewSession", func(engine *xorm.Engine) (s *xorm.Session) {
			return &session
		})
		err := client.BatchCreatePipelineTasks(tasks)

		var preTaskId = uint64(0)
		for i := data.num - 1; i >= 0; i-- {
			assert.NotZero(t, tasks[i].ID)
			if preTaskId == 0 {
				preTaskId = tasks[i].ID
			} else {
				assert.Equal(t, preTaskId, tasks[i].ID+1, fmt.Sprintf("index %v", i))
			}
			preTaskId = tasks[i].ID
		}
		assert.NoError(t, err)
		patch.Unpatch()
		patch1.Unpatch()
		patch2.Unpatch()
		patch3.Unpatch()
		patch4.Unpatch()
		patch5.Unpatch()
	}
}

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
