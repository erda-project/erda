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
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
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
					rlog.PInfof(pipelineID, "add into queue, waiting to pop from the queue")
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
					<-popCh
					rlog.PInfof(pipelineID, "pop from the queue, begin reconcile")

					// construct context for pipeline reconciler
					pCtx := makeContextForPipelineReconcile(pipelineID)

					// reconciler
					reconcileErr := r.reconcile(pCtx, pipelineID)

					// if reconcile failed, put and wait next reconcile
					if reconcileErr != nil {
						rlog.PErrorf(pipelineID, "failed to reconcile pipeline, err: %v", err)
						r.reconcileAgain(pipelineID)
						return
					}
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
