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
