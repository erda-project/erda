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
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/statusutil"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler/legacy/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

func (r *Reconciler) ReconcileOnePipelineUntilDone(ctx context.Context, pipelineID uint64) {
	// there may be multiple reconciles here, and subsequent reconciles will cause the pipeline to succeed directly
	_, ok := r.processingPipelines.Load(pipelineID)
	if ok {
		rlog.PErrorf(pipelineID, "pipeline duplication reconcile")
		return
	} else {
		r.processingPipelines.Store(pipelineID, pipelineID)
		defer func() {
			r.processingPipelines.Delete(pipelineID)
		}()
	}

	var p spec.Pipeline
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		p, err = r.dbClient.GetPipeline(pipelineID)
		if err != nil {
			rlog.PWarnf(pipelineID, "failed to get pipeline, err: %v, will continue until get pipeline", err)
			return false, err
		}
		return true, nil
	})
	if p.Status.IsEndStatus() {
		rlog.PErrorf(pipelineID, "unable to reconcile end status pipeline")
		return
	}

	if err := r.updateStatusBeforeReconcile(p); err != nil {
		rlog.PErrorf(p.ID, "Failed to update pipeline status before reconcile, err: %v", err)
		return
	}

	// construct context for pipeline reconciler
	pCtx := makeContextForPipelineReconcile(pipelineID)
	go func() {
		for {
			select {
			case <-ctx.Done():
				pCancel, ok := pCtx.Value(ctxKeyPipelineExitChCancelFunc).(context.CancelFunc)
				if ok {
					pCancel()
				}
				return
			case <-pCtx.Done():
				return
			}
		}
	}()

	// continue reconcile
	for {
		reconcileErr := r.internalReconcileOnePipeline(pCtx, pipelineID)
		if reconcileErr == nil {
			break
		}
		rlog.PErrorf(pipelineID, "failed to reconcile pipeline(auto retry), err: %v", reconcileErr)
		time.Sleep(time.Second * 5)
		continue
	}

	// teardown
	r.teardownCurrentReconcile(pCtx, pipelineID)

	// update status
	var pipelineWithTasks *spec.PipelineWithTasks
	for {
		p, err := r.updateStatusAfterReconcile(pCtx, pipelineID)
		if err == nil {
			pipelineWithTasks = p
			break
		}
		rlog.PErrorf(pipelineID, "failed to update status after reconcile(auto retry), err: %v", err)
		time.Sleep(time.Second * 5)
		continue
	}

	// teardown
	r.teardownPipeline(ctx, pipelineWithTasks)
}

// updateStatusBeforeReconcile update pipeline status to running
func (r *Reconciler) updateStatusBeforeReconcile(p spec.Pipeline) error {
	if !p.Status.IsRunningStatus() {
		oldStatus := p.Status
		p.Status = apistructs.PipelineStatusRunning
		if err := r.UpdatePipelineStatus(&p); err != nil {
			return err
		}
		rlog.PInfof(p.ID, "update pipeline status (%s -> %s)", oldStatus, apistructs.PipelineStatusRunning)
	}
	return nil
}

// updateStatusAfterReconcile get latest pipeline after reconcile
func (r *Reconciler) updateStatusAfterReconcile(ctx context.Context, pipelineID uint64) (*spec.PipelineWithTasks, error) {
	pipelineWithTasks, err := r.dbClient.GetPipelineWithTasks(pipelineID)
	if err != nil {
		rlog.PErrorf(pipelineID, "failed to get pipeline with tasks, err: %v", err)
		return nil, err
	}

	p := pipelineWithTasks.Pipeline
	tasks := pipelineWithTasks.Tasks
	// if status is end status like stopByUser, should return immediately
	if p.Status.IsEndStatus() {
		return pipelineWithTasks, nil
	}

	// calculate pipeline status by tasks
	calcPStatus := statusutil.CalculatePipelineStatusV2(tasks)
	if calcPStatus != p.Status {
		oldStatus := p.Status
		p.Status = calcPStatus
		if err := r.UpdatePipelineStatus(p); err != nil {
			return nil, err
		}
		rlog.PInfof(p.ID, "update pipeline status (%s -> %s)", oldStatus, calcPStatus)
	}
	return pipelineWithTasks, nil
}

// makeContextForPipelineReconcile
func makeContextForPipelineReconcile(pipelineID uint64) context.Context {
	pCtx := context.WithValue(context.Background(), ctxKeyPipelineID, pipelineID)
	exitCh := make(chan struct{})
	pCtx = context.WithValue(pCtx, ctxKeyPipelineExitCh, exitCh)
	pCtx, pCancel := context.WithCancel(pCtx)
	pCtx = context.WithValue(pCtx, ctxKeyPipelineExitChCancelFunc, pCancel)
	go func() {
		select {
		case <-exitCh:
			// default receiver to prevent send block
		}
	}()
	return pCtx
}

// parsePipelineIDFromWatchedKey get pipelineID from watched key.
func parsePipelineIDFromWatchedKey(key string) (uint64, error) {
	pipelineIDStr := strutil.TrimPrefixes(key, etcdReconcilerWatchPrefix)
	return strconv.ParseUint(pipelineIDStr, 10, 64)
}

// makePipelineWatchedKey construct etcd watched key by pipelineID.
func makePipelineWatchedKey(pipelineID uint64) string {
	return fmt.Sprintf("%s%d", etcdReconcilerWatchPrefix, pipelineID)
}
