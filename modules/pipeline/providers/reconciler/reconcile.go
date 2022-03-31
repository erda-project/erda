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
	"runtime/debug"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/rutil"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/schedulabletask"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (r *provider) ReconcileOnePipeline(ctx context.Context, pipelineID uint64) {
	// recover
	defer func() {
		if err := recover(); err != nil {
			r.Log.Errorf("panic while reconcile one pipeline until done, cancel reconcile, pipelineID: %d, err: %v", pipelineID, err)
			debug.PrintStack()
		}
	}()

	// make pipelineID-level context to support cancel reconcile
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// generate pipeline-id-level pr
	pr := r.generatePipelineReconcilerForEachPipelineID()

	// reconcile
	rutil.ContinueWorking(ctx, r.Log, func(ctx context.Context) (waitDuration rutil.WaitDuration) {

		// fetch pipeline detail
		p := r.mustFetchPipelineDetail(ctx, pipelineID)

		// TODO handle outer stop at reconciler side later
		if p.Status == apistructs.PipelineStatusStopByUser {
			// teardown
			pr.TeardownAfterReconcileDone(ctx, p)
			return rutil.ContinueWorkingAbort
		}

		// check need reconcile
		if !pr.NeedReconcile(ctx, p) {
			return rutil.ContinueWorkingAbort
		}

		// prepare before reconcile
		if err := pr.PrepareBeforeReconcile(ctx, p); err != nil {
			r.Log.Errorf("failed to prepare before reconcile(auto retry), pipelineID: %d, err: %v", p.ID, err)
			return rutil.ContinueWorkingWithDefaultInterval
		}

		// get tasks which can be scheduled
		schedulableTasks, err := pr.GetTasksCanBeConcurrentlyScheduled(ctx, p)
		//defer pr.releaseTasksCanBeConcurrentlyScheduled(ctx, p, schedulableTasks)
		if err != nil {
			r.Log.Errorf("failed to get tasks can be concurrently scheduled(auto retry), pipelineID: %d, err: %v", p.ID, err)
			return rutil.ContinueWorkingWithDefaultInterval
		}

		// reconcile schedulable tasks
		if err := pr.ReconcileSchedulableTasks(ctx, p, schedulableTasks); err != nil {
			r.Log.Errorf("failed to reconcile one pipeline schedulable tasks(auto retry), pipelineID: %d, err: %v", p.ID, err)
			return rutil.ContinueWorkingWithDefaultInterval
		}

		// calculate and update current reconcile status
		if err := pr.UpdateCurrentReconcileStatusIfNecessary(ctx, p); err != nil {
			r.Log.Errorf("failed to calculate pipeline status(auto retry), pipelineID: %d, err: %v", p.ID, err)
			return rutil.ContinueWorkingWithDefaultInterval
		}

		// check pipeline reconcile done
		done := pr.IsReconcileDone(ctx, p)
		if !done {
			// not done, enter into next reconcile immediately
			return rutil.ContinueWorkingImmediately
		}
		// done, do teardown and exit loop
		pr.TeardownAfterReconcileDone(ctx, p)

		// all done, exit
		return rutil.ContinueWorkingAbort

	}, rutil.WithContinueWorkingDefaultRetryInterval(r.Cfg.RetryInterval))
}

func (r *provider) generatePipelineReconcilerForEachPipelineID() *defaultPipelineReconciler {
	pr := &defaultPipelineReconciler{
		log:                  r.Log.Sub("pipeline"),
		st:                   &schedulabletask.DagImpl{},
		resourceGC:           r.ResourceGC,
		cronCompensator:      r.CronCompensator,
		r:                    r,
		dbClient:             r.dbClient,
		processingTasks:      sync.Map{},
		defaultRetryInterval: r.Cfg.RetryInterval,
		calculatedPipelineStatusByAllReconciledTasks: "",
	}
	return pr
}

func (pr *defaultPipelineReconciler) releaseTasksCanBeConcurrentlyScheduled(ctx context.Context, p *spec.Pipeline, tasks []*spec.PipelineTask) {
	for _, task := range tasks {
		pr.processingTasks.Delete(task.NodeName())
	}
}
