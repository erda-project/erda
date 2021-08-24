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
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

// Listen watch incoming pipelines which need to be scheduled from etcd.
func (r *Reconciler) Listen() {
	rlog.Infof("start listen")
	for {
		_ = r.js.IncludeWatch().Watch(context.Background(), etcdReconcilerWatchPrefix, true, true, true, nil,
			func(key string, _ interface{}, t storetypes.ChangeType) (_ error) {

				// async reconcile, non-blocking, so we can watch subsequent incoming pipelines
				go func() {
					rlog.Infof("watched a key change, key: %s, changeType: %s", key, t.String())

					// parse pipelineID
					pipelineID, err := parsePipelineIDFromWatchedKey(key)
					if err != nil {
						rlog.Errorf("failed to parse pipelineID from watched key, key: %s, err: %v", key, err)
						return
					}

					// add into queue
					popCh, needRetryIfErr, err := r.QueueManager.PutPipelineIntoQueue(pipelineID)
					if err != nil {
						rlog.PErrorf(pipelineID, "failed to put pipeline into queue")
						if needRetryIfErr {
							r.reconcileAgain(pipelineID)
							return
						}
						// no need retry, treat as failed
						_ = r.updatePipelineStatus(&spec.Pipeline{
							PipelineBase: spec.PipelineBase{
								ID:     pipelineID,
								Status: apistructs.PipelineStatusFailed,
							}})
						return
					}
					rlog.PInfof(pipelineID, "added into queue, waiting to pop from the queue")
					<-popCh
					rlog.PInfof(pipelineID, "pop from the queue, begin reconcile")
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

					// reconciler
					reconcileErr := r.reconcile(pCtx, pipelineID)

					defer func() {
						r.teardownCurrentReconcile(pCtx, pipelineID)
						if err := r.updateStatusAfterReconcile(pCtx, pipelineID); err != nil {
							rlog.PErrorf(pipelineID, "failed to update status after reconcile, err: %v", err)
						}

						// if reconcile failed, put and wait next reconcile
						if reconcileErr != nil {
							rlog.PErrorf(pipelineID, "failed to reconcile pipeline, err: %v", err)
							r.reconcileAgain(pipelineID)
						}
					}()
				}()

				return nil
			},
		)
	}
}

// reconcileAgain add to reconciler again, wait next reconcile
func (r *Reconciler) reconcileAgain(pipelineID uint64) {
	time.Sleep(time.Second * 5)
	r.Add(pipelineID)
}

// updateStatusBeforeReconcile update pipeline status to running
func (r *Reconciler) updateStatusBeforeReconcile(p spec.Pipeline) error {
	if !p.Status.IsRunningStatus() {
		oldStatus := p.Status
		p.Status = apistructs.PipelineStatusRunning
		if err := r.updatePipelineStatus(&p); err != nil {
			return err
		}
		rlog.PInfof(p.ID, "update pipeline status (%s -> %s)", oldStatus, apistructs.PipelineStatusRunning)
	}
	return nil
}

// updateStatusAfterReconcile get latest pipeline after reconcile
func (r *Reconciler) updateStatusAfterReconcile(ctx context.Context, pipelineID uint64) error {
	pipelineWithTasks, err := r.dbClient.GetPipelineWithTasks(pipelineID)
	if err != nil {
		rlog.PErrorf(pipelineID, "failed to get pipeline with tasks, err: %v", err)
		return err
	}
	defer r.teardownPipeline(ctx, pipelineWithTasks)

	p := pipelineWithTasks.Pipeline
	tasks := pipelineWithTasks.Tasks
	// if status is end status like stopByUser, should return immediately
	if p.Status.IsEndStatus() {
		return nil
	}

	// calculate pipeline status by tasks
	calcPStatus := statusutil.CalculatePipelineStatusV2(tasks)
	if calcPStatus != p.Status {
		oldStatus := p.Status
		p.Status = calcPStatus
		if err := r.updatePipelineStatus(p); err != nil {
			return err
		}
		rlog.PInfof(p.ID, "update pipeline status (%s -> %s)", oldStatus, calcPStatus)
	}
	return nil
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
