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
	"context"
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func Test_defaultPipelineReconciler_IsReconcileDone(t *testing.T) {
	p := &spec.Pipeline{
		PipelineBase: spec.PipelineBase{
			Status: apistructs.PipelineStatusRunning,
		},
	}
	pr := &defaultPipelineReconciler{}
	ctx := context.TODO()

	// pipeline is canceled
	// tasks: s1c1, s2c1
	pr.flagCanceling = true
	pr.calculatedStatusForTaskUse = apistructs.PipelineStatusStopByUser
	pr.totalTaskNumber = &[]int{2}[0]
	pr.processedTasks.Store("s1c1", struct{}{})
	pr.processingTasks.Store("s2c1", struct{}{})
	done := pr.IsReconcileDone(ctx, p)
	if !done {
		t.Fatalf("should done")
	}

	// pipeline running
	// tasks: s1c1(done), s2c1(done), s2c2(analyzed)
	pr = &defaultPipelineReconciler{}
	pr.flagCanceling = false
	pr.totalTaskNumber = &[]int{3}[0]
	pr.processedTasks.Store("s1c1", struct{}{})
	pr.processedTasks.Store("s2c1", struct{}{})
	done = pr.IsReconcileDone(ctx, p)
	if done {
		t.Fatalf("should running")
	}

	// pipeline all tasks done
	// tasks: s1c1(done), s2c1(done), s2c2(failed)
	pr = &defaultPipelineReconciler{}
	pr.flagCanceling = false
	pr.totalTaskNumber = &[]int{3}[0]
	pr.processedTasks.Store("s1c1", struct{}{})
	pr.processedTasks.Store("s2c1", struct{}{})
	pr.processedTasks.Store("s2c2", struct{}{})
	done = pr.IsReconcileDone(ctx, p)
	if !done {
		t.Fatalf("should done")
	}
}
