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

package reconciler

import (
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func TestResetTaskForAbnormalRetry(t *testing.T) {
	tr := &taskrun.TaskRun{
		P: &spec.Pipeline{},
		Task: &spec.PipelineTask{
			Status: apistructs.PipelineStatusAnalyzeFailed,
		},
	}
	m := monkey.PatchInstanceMethod(reflect.TypeOf(tr), "Update",
		func(tr *taskrun.TaskRun) {
			return
		})
	defer m.Unpatch()
	tm := monkey.Patch(time.Sleep, func(d time.Duration) {
		return
	})
	defer tm.Unpatch()
	resetTaskForAbnormalRetry(tr, 1)
	assert.Equal(t, apistructs.PipelineStatusAnalyzeFailed, tr.Task.Status)
}
