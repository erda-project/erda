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
