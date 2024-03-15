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

package dbclient

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"xorm.io/xorm/names"

	"github.com/erda-project/erda-infra/providers/mysqlxorm/sqlite3"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
)

const (
	dbSourceName = "test-*.db"
	mode         = "rwc"
)

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

func TestUpdatePipelineTaskTime(t *testing.T) {
	dbname := filepath.Join(os.TempDir(), dbSourceName)

	sqlite3Db, err := sqlite3.NewSqlite3(dbname+"?mode="+mode, sqlite3.WithJournalMode(sqlite3.MEMORY), sqlite3.WithRandomName(true))
	defer func() {
		if sqlite3Db != nil {
			defer sqlite3Db.Close()
		}
	}()
	if err != nil {
		panic(err)
	}

	sqlite3Db.DB().SetMapper(names.GonicMapper{})
	err = sqlite3Db.DB().Sync2(&spec.PipelineTask{})
	if err != nil {
		panic(err)
	}

	client := Client{
		Engine: sqlite3Db.DB(),
	}

	curDate := time.Date(2024, 3, 12, 17, 25, 0, 0, time.Now().Location())

	// insert task
	tasks := []*spec.PipelineTask{
		{ID: 1, PipelineID: 1, CostTimeSec: 10, TimeBegin: curDate, TimeEnd: curDate},
		{ID: 2, PipelineID: 2, CostTimeSec: 10, TimeBegin: curDate, TimeEnd: curDate},
		{ID: 3, PipelineID: 3, CostTimeSec: 10, TimeBegin: curDate, TimeEnd: curDate},
	}

	for _, task := range tasks {
		err = client.CreatePipelineTask(task)
		if err != nil {
			panic(err)
		}
	}

	type Want struct {
		Err          bool
		pipelineTask *spec.PipelineTask
	}

	addTime := curDate.Add(time.Hour)
	one := uint64(1)

	testCase := []struct {
		desc     string
		pipeline *spec.Pipeline
		want     Want
	}{
		{
			desc: "parent task id is null",
			pipeline: &spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					TimeBegin:    &curDate,
					TimeEnd:      &curDate,
					CostTimeSec:  20,
					ParentTaskID: nil,
				},
			},
			want: Want{
				Err: true,
			},
		},
		{
			desc: "normal",
			pipeline: &spec.Pipeline{
				PipelineBase: spec.PipelineBase{
					TimeBegin:    &addTime,
					TimeEnd:      &addTime,
					CostTimeSec:  20,
					ParentTaskID: &one,
				},
			},
			want: Want{
				Err: false,
				pipelineTask: &spec.PipelineTask{
					ID:          1,
					TimeBegin:   addTime,
					TimeEnd:     addTime,
					CostTimeSec: 20,
				},
			},
		},
	}

	for _, tt := range testCase {
		t.Log(tt.desc)
		err = client.UpdatePipelineTaskTime(tt.pipeline)
		if tt.want.Err {
			assert.NotNil(t, err)
			continue
		}

		task, _ := client.GetPipelineTask(tt.want.pipelineTask.ID)
		assert.Equal(t, tt.want.pipelineTask.ID, task.ID)
		assert.Equal(t, tt.want.pipelineTask.TimeBegin, task.TimeBegin)
		assert.Equal(t, tt.want.pipelineTask.TimeEnd, task.TimeEnd)
		assert.Equal(t, tt.want.pipelineTask.CostTimeSec, task.CostTimeSec)
	}
}
