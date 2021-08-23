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

//import (
//	"sync"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/modules/pipeline/spec"
//)
//
//func TestGetSchedulableTasks(t *testing.T) {
//	r := Reconciler{
//		processingTasks: sync.Map{},
//	}
//	tests := []struct {
//		name  string
//		p     *spec.Pipeline
//		tasks []*spec.PipelineTask
//
//		expectSchedulableLen int
//	}{
//		{
//			name: "only t1",
//			p:    &spec.Pipeline{PipelineBase: spec.PipelineBase{ID: 1}},
//			tasks: []*spec.PipelineTask{
//				{Name: "t1", Status: apistructs.PipelineStatusAnalyzed, Extra: spec.PipelineTaskExtra{RunAfter: []string{}}},
//			},
//			expectSchedulableLen: 1,
//		},
//		{
//			name: "t1 done, t2 depends on t1",
//			p:    &spec.Pipeline{PipelineBase: spec.PipelineBase{ID: 1}},
//			tasks: []*spec.PipelineTask{
//				{Name: "t1", Status: apistructs.PipelineStatusSuccess, Extra: spec.PipelineTaskExtra{RunAfter: []string{}}},
//				{Name: "t2", Status: apistructs.PipelineStatusAnalyzed, Extra: spec.PipelineTaskExtra{RunAfter: []string{"t1"}}},
//			},
//			expectSchedulableLen: 1,
//		},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			schedulableTasks, err := r.getSchedulableTasks(test.p, test.tasks)
//			assert.NoError(t, err)
//			assert.Equal(t, test.expectSchedulableLen, len(schedulableTasks))
//		})
//	}
//}
