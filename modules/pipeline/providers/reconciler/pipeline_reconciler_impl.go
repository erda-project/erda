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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/statusutil"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// updateCalculatedPipelineStatusForTaskUseField by:
// 1. other flags (higher priority)
// 2. all reconciled tasks
func (pr *defaultPipelineReconciler) updateCalculatedPipelineStatusForTaskUseField(ctx context.Context, p *spec.Pipeline) error {
	newStatus := p.Status
	defer func() {
		pr.calculatedStatusForTaskUse = newStatus
	}()

	// check if done
	if newStatus.IsEndStatus() {
		return nil
	}

	// check flags
	if pr.flagCanceling {
		newStatus = apistructs.PipelineStatusStopByUser
		return nil
	}

	// calculate new status
	calculatedStatus, err := pr.calculatePipelineStatusForTaskUseField(ctx, p)
	if err != nil {
		pr.log.Errorf("failed to calculate new status by reconciled tasks(auto retry), pipelineID: %d, err: %v", p.ID, err)
		return err
	}
	newStatus = calculatedStatus
	return nil
}

func (pr *defaultPipelineReconciler) calculatePipelineStatusForTaskUseField(ctx context.Context, p *spec.Pipeline) (apistructs.PipelineStatus, error) {
	// get all reconciled tasks
	reconciledDBTasks, err := pr.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return "", err
	}
	var reconciledTasks []*spec.PipelineTask
	for _, t := range reconciledDBTasks {
		t := t
		reconciledTasks = append(reconciledTasks, &t)
	}

	// calculate new pipeline status
	calculatedPipelineStatusByAllReconciledTasks := statusutil.CalculatePipelineStatusV2(reconciledTasks)
	// consider some special cases:
	// - no reconciled tasks but pipeline actually have tasks (to resolve first time of loop)
	if calculatedPipelineStatusByAllReconciledTasks.IsSuccessStatus() && len(reconciledTasks) == 0 && *pr.totalTaskNumber > 0 {
		calculatedPipelineStatusByAllReconciledTasks = apistructs.PipelineStatusRunning
	}

	// update status
	return calculatedPipelineStatusByAllReconciledTasks, nil
}

func (pr *defaultPipelineReconciler) getCalculatedStatusByAllReconciledTasks() apistructs.PipelineStatus {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	return pr.calculatedStatusForTaskUse
}

func (pr *defaultPipelineReconciler) getFlagCanceling() bool {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	return pr.flagCanceling
}

func (pr *defaultPipelineReconciler) setTotalTaskNumber(num int) {
	pr.lock.Lock()
	defer pr.lock.Unlock()
	pr.totalTaskNumber = &num
}

func (pr *defaultPipelineReconciler) setTotalTaskNumberBeforeReconcilePipeline(ctx context.Context, p *spec.Pipeline) error {
	allTasks, err := pr.r.YmlTaskMergeDBTasks(p)
	if err != nil {
		return err
	}
	// set processed tasks
	for _, task := range allTasks {
		task := task
		if task.Status.IsEndStatus() || task.Status.IsDisabledStatus() {
			pr.processedTasks.Store(task.NodeName(), struct{}{})
		}
	}
	pr.setTotalTaskNumber(len(allTasks))
	return nil
}
