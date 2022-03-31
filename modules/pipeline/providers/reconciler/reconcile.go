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

	"github.com/erda-project/erda-infra/pkg/safe"
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

	// fetch pipeline detail
	p := r.mustFetchPipelineDetail(ctx, pipelineID)

	// TODO handle outer stop at reconciler side later
	if p.Status == apistructs.PipelineStatusStopByUser {
		// teardown
		pr.TeardownAfterReconcileDone(ctx, p)
		return
	}

	// check need reconcile
	if !pr.NeedReconcile(ctx, p) {
		return
	}

	// prepare before reconcile
	pr.PrepareBeforeReconcile(ctx, p)

	// continue calculate schedulable tasks
	safe.Go(func() { pr.continuePushSchedulableTasks(ctx, p) })

	// continue reconcile schedulable tasks
	safe.Go(func() { pr.continueScheduleTasks(ctx, p) })

	// wait pipeline done and do the teardown
	safe.Do(func() { pr.waitPipelineDoneAndDoTeardown(ctx, p) })
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
		chanToTriggerNextLoop:                        make(chan struct{}),
		schedulableTaskChan:                          make(chan *spec.PipelineTask),
		doneChan:                                     make(chan struct{}),
	}
	return pr
}

func (pr *defaultPipelineReconciler) releaseTasksCanBeConcurrentlyScheduled(ctx context.Context, p *spec.Pipeline, tasks []*spec.PipelineTask) {
	for _, task := range tasks {
		pr.processingTasks.Delete(task.NodeName())
	}
}

func (pr *defaultPipelineReconciler) waitPipelineDoneAndDoTeardown(ctx context.Context, p *spec.Pipeline) {
	select {
	case <-ctx.Done():
		return
	case <-pr.doneChan:
		pr.TeardownAfterReconcileDone(ctx, p)
		return
	}
}

func (pr *defaultPipelineReconciler) continuePushSchedulableTasks(ctx context.Context, p *spec.Pipeline) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pr.chanToTriggerNextLoop:
			rutil.ContinueWorking(ctx, pr.log, func(ctx context.Context) rutil.WaitDuration {
				// set calculated pipeline status by all reconciled tasks from db
				if err := pr.calculatePipelineStatusByAllReconciledTasks(ctx, p); err != nil {
					pr.log.Errorf("failed to calculate pipeline status by all reconciled tasks(auto retry), pipelineID: %d, err: %v", p.ID, err)
					return rutil.ContinueWorkingWithDefaultInterval
				}
				schedulableTasks, err := pr.GetTasksCanBeConcurrentlyScheduled(ctx, p)
				if err != nil {
					pr.log.Errorf("failed to get tasks can be concurrently scheduled(auto retry), pipelineID: %d, err: %v", p.ID, err)
					return rutil.ContinueWorkingWithDefaultInterval
				}
				for _, task := range schedulableTasks {
					pr.schedulableTaskChan <- task
				}
				// calculate if all done
				if len(schedulableTasks) == 0 {
					// calculate and update current reconcile status
					if err := pr.UpdateCurrentReconcileStatusIfNecessary(ctx, p); err != nil {
						pr.log.Errorf("failed to calculate pipeline status(auto retry), pipelineID: %d, err: %v", p.ID, err)
						return rutil.ContinueWorkingWithDefaultInterval
					}

					// check pipeline reconcile done
					done := pr.IsReconcileDone(ctx, p)
					if done && pr.doneChan != nil {
						pr.doneChan <- struct{}{}
						close(pr.doneChan)
						pr.doneChan = nil // set doneChan to nil to guarantee only teardown once
						return rutil.ContinueWorkingAbort
					}
				}
				return rutil.ContinueWorkingAbort
			}, rutil.WithContinueWorkingDefaultRetryInterval(pr.defaultRetryInterval))
		}
	}
}

func (pr *defaultPipelineReconciler) continueScheduleTasks(ctx context.Context, p *spec.Pipeline) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-pr.schedulableTaskChan:
			if !ok {
				return
			}
			safe.Go(func() { pr.ReconcileOneSchedulableTask(ctx, p, task) })
		}
	}
}
