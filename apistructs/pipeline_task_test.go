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

package apistructs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipelineTaskLoop_Duplicate(t *testing.T) {
	var l *PipelineTaskLoop
	fmt.Println(l.Duplicate())
}

func TestPipelineTaskAppendError(t *testing.T) {
	task := PipelineTaskDTO{}
	task.Result.Errors = task.Result.AppendError(&PipelineTaskErrResponse{Msg: "a"})
	task.Result.Errors = task.Result.AppendError(&PipelineTaskErrResponse{Msg: "a"})
	assert.Equal(t, 1, len(task.Result.Errors))
	task.Result.Errors = task.Result.AppendError(&PipelineTaskErrResponse{Msg: "b"})
	assert.Equal(t, 2, len(task.Result.Errors))
	startA := time.Date(2021, 8, 19, 10, 10, 0, 0, time.Local)
	endA := time.Date(2021, 8, 19, 10, 30, 0, 0, time.Local)
	task.Result.Errors = task.Result.AppendError(&PipelineTaskErrResponse{Msg: "a", Ctx: PipelineTaskErrCtx{StartTime: startA, EndTime: endA}})
	assert.Equal(t, 3, len(task.Result.Errors))
	start := time.Date(2021, 8, 19, 10, 9, 0, 0, time.Local)
	end := time.Date(2021, 8, 19, 10, 29, 0, 0, time.Local)
	task.Result.Errors = task.Result.AppendError(&PipelineTaskErrResponse{Msg: "a", Ctx: PipelineTaskErrCtx{StartTime: start, EndTime: end}})
	assert.Equal(t, uint64(2), task.Result.Errors[2].Ctx.Count)
	assert.Equal(t, 3, len(task.Result.Errors))
	assert.Equal(t, start.Unix(), task.Result.Errors[2].Ctx.StartTime.Unix())
	assert.Equal(t, endA.Unix(), task.Result.Errors[2].Ctx.EndTime.Unix())
}

func TestConvertErrors(t *testing.T) {
	task := PipelineTaskDTO{}
	start := time.Date(2021, 8, 24, 9, 45, 1, 1, time.Local)
	end := time.Date(2021, 8, 24, 9, 46, 1, 1, time.Local)
	task.Result.Errors = task.Result.AppendError(&PipelineTaskErrResponse{Msg: "err", Ctx: PipelineTaskErrCtx{
		StartTime: start,
		EndTime:   end,
		Count:     2,
	}})
	task.Result.ConvertErrors()
	assert.Equal(t, fmt.Sprintf("err\nstartTime: %s\nendTime: %s\ncount: %d", start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05"), 2), task.Result.Errors[0].Msg)
}
